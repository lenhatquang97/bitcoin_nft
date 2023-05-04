package wallet

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	psbtMagicLength = 5
)

var (
	// psbtMagic is the separator.
	psbtMagic = [psbtMagicLength]byte{0x70,
		0x73, 0x62, 0x74, 0xff, // = "psbt" + 0xff sep
	}
)

// Serialize creates a binary serialization of the referenced Packet struct
// with lexicographical ordering (by key) of the subsections.
func SerializePacket(p *psbt.Packet, w io.Writer) error {
	// First we write out the precise set of magic bytes that identify a
	// valid PSBT transaction.
	if _, err := w.Write(psbtMagic[:]); err != nil {
		return err
	}

	// Next we prep to write out the unsigned transaction by first
	// serializing it into an intermediate buffer.
	serializedTx := bytes.NewBuffer(
		make([]byte, 0, p.UnsignedTx.SerializeSize()),
	)
	if err := p.UnsignedTx.Serialize(serializedTx); err != nil {
		return err
	}

	// Now that we have the serialized transaction, we'll write it out to
	// the proper global type.
	err := serializeKVPairWithType(
		w, uint8(psbt.UnsignedTxType), nil, serializedTx.Bytes(),
	)
	if err != nil {
		return err
	}

	// Unknown is a special case; we don't have a key type, only a key and
	// a value field
	for _, kv := range p.Unknowns {
		err := serializeKVpair(w, kv.Key, kv.Value)
		if err != nil {
			return err
		}
	}

	// With that our global section is done, so we'll write out the
	// separator.
	separator := []byte{0x00}
	if _, err := w.Write(separator); err != nil {
		return err
	}

	for _, pInput := range p.Inputs {
		err := SerializeCustomInput(&pInput, w)
		if err != nil {
			return err
		}

		if _, err := w.Write(separator); err != nil {
			return err
		}
	}

	for _, pOutput := range p.Outputs {
		err := SerializeCustomOutput(&pOutput, w)
		if err != nil {
			return err
		}

		if _, err := w.Write(separator); err != nil {
			return err
		}
	}

	return nil
}

func serializeKVPairWithType(w io.Writer, kt uint8, keydata []byte,
	value []byte) error {

	// If the key has no data, then we write a blank slice.
	if keydata == nil {
		keydata = []byte{}
	}

	// The final key to be written is: {type} || {keyData}
	serializedKey := append([]byte{kt}, keydata...)
	return serializeKVpair(w, serializedKey, value)
}

func serializeKVpair(w io.Writer, key []byte, value []byte) error {
	if err := wire.WriteVarBytes(w, 0, key); err != nil {
		return err
	}

	return wire.WriteVarBytes(w, 0, value)
}

// serialize attempts to serialize the target PInput into the passed io.Writer.
func SerializeCustomInput(pi *psbt.PInput, w io.Writer) error {
	if !pi.IsSane() {
		return psbt.ErrInvalidPsbtFormat
	}

	if pi.NonWitnessUtxo != nil {
		var buf bytes.Buffer
		err := pi.NonWitnessUtxo.Serialize(&buf)
		if err != nil {
			return err
		}

		err = serializeKVPairWithType(
			w, uint8(psbt.NonWitnessUtxoType), nil, buf.Bytes(),
		)
		if err != nil {
			return err
		}
	}
	if pi.WitnessUtxo != nil {
		var buf bytes.Buffer
		err := wire.WriteTxOut(&buf, 0, 0, pi.WitnessUtxo)
		if err != nil {
			return err
		}

		err = serializeKVPairWithType(
			w, uint8(psbt.WitnessUtxoType), nil, buf.Bytes(),
		)
		if err != nil {
			return err
		}
	}

	if pi.FinalScriptSig == nil && pi.FinalScriptWitness == nil {
		sort.Sort(psbt.PartialSigSorter(pi.PartialSigs))
		for _, ps := range pi.PartialSigs {
			err := serializeKVPairWithType(
				w, uint8(psbt.PartialSigType), ps.PubKey,
				ps.Signature,
			)
			if err != nil {
				return err
			}
		}

		if pi.SighashType != 0 {
			var shtBytes [4]byte
			binary.LittleEndian.PutUint32(
				shtBytes[:], uint32(pi.SighashType),
			)

			err := serializeKVPairWithType(
				w, uint8(psbt.SighashType), nil, shtBytes[:],
			)
			if err != nil {
				return err
			}
		}

		if pi.RedeemScript != nil {
			err := serializeKVPairWithType(
				w, uint8(psbt.RedeemScriptInputType), nil,
				pi.RedeemScript,
			)
			if err != nil {
				return err
			}
		}

		if pi.WitnessScript != nil {
			err := serializeKVPairWithType(
				w, uint8(psbt.WitnessScriptInputType), nil,
				pi.WitnessScript,
			)
			if err != nil {
				return err
			}
		}

		sort.Sort(psbt.Bip32Sorter(pi.Bip32Derivation))
		for _, kd := range pi.Bip32Derivation {
			err := serializeKVPairWithType(
				w,
				uint8(psbt.Bip32DerivationInputType), kd.PubKey,
				psbt.SerializeBIP32Derivation(
					kd.MasterKeyFingerprint, kd.Bip32Path,
				),
			)
			if err != nil {
				return err
			}
		}

		if pi.TaprootKeySpendSig != nil {
			err := serializeKVPairWithType(
				w, uint8(psbt.TaprootKeySpendSignatureType), nil,
				pi.TaprootKeySpendSig,
			)
			if err != nil {
				return err
			}
		}

		sort.Slice(pi.TaprootScriptSpendSig, func(i, j int) bool {
			return pi.TaprootScriptSpendSig[i].SortBefore(
				pi.TaprootScriptSpendSig[j],
			)
		})
		for _, scriptSpend := range pi.TaprootScriptSpendSig {
			keyData := append([]byte{}, scriptSpend.XOnlyPubKey...)
			keyData = append(keyData, scriptSpend.LeafHash...)
			value := append([]byte{}, scriptSpend.Signature...)
			if scriptSpend.SigHash != txscript.SigHashDefault {
				value = append(value, byte(scriptSpend.SigHash))
			}
			err := serializeKVPairWithType(
				w, uint8(psbt.TaprootScriptSpendSignatureType),
				keyData, value,
			)
			if err != nil {
				return err
			}
		}

		sort.Slice(pi.TaprootLeafScript, func(i, j int) bool {
			return pi.TaprootLeafScript[i].SortBefore(
				pi.TaprootLeafScript[j],
			)
		})
		for _, leafScript := range pi.TaprootLeafScript {
			value := append([]byte{}, leafScript.Script...)
			value = append(value, byte(leafScript.LeafVersion))
			err := serializeKVPairWithType(
				w, uint8(psbt.TaprootLeafScriptType),
				leafScript.ControlBlock, value,
			)
			if err != nil {
				return err
			}
		}

		sort.Slice(pi.TaprootBip32Derivation, func(i, j int) bool {
			return pi.TaprootBip32Derivation[i].SortBefore(
				pi.TaprootBip32Derivation[j],
			)
		})
		for _, derivation := range pi.TaprootBip32Derivation {
			value, err := psbt.SerializeTaprootBip32Derivation(
				derivation,
			)
			if err != nil {
				return err
			}
			err = serializeKVPairWithType(
				w, uint8(psbt.TaprootBip32DerivationInputType),
				derivation.XOnlyPubKey, value,
			)
			if err != nil {
				return err
			}
		}

		if pi.TaprootInternalKey != nil {
			err := serializeKVPairWithType(
				w, uint8(psbt.TaprootInternalKeyInputType), nil,
				pi.TaprootInternalKey,
			)
			if err != nil {
				return err
			}
		}

		if pi.TaprootMerkleRoot != nil {
			err := serializeKVPairWithType(
				w, uint8(psbt.TaprootMerkleRootType), nil,
				pi.TaprootMerkleRoot,
			)
			if err != nil {
				return err
			}
		}
	}

	if pi.FinalScriptSig != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.FinalScriptSigType), nil, pi.FinalScriptSig,
		)
		if err != nil {
			return err
		}
	}

	if pi.FinalScriptWitness != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.FinalScriptWitnessType), nil, pi.FinalScriptWitness,
		)
		if err != nil {
			return err
		}
	}

	// Unknown is a special case; we don't have a key type, only a key and
	// a value field.
	for _, kv := range pi.Unknowns {
		err := serializeKVpair(w, kv.Key, kv.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func SerializeCustomOutput(po *psbt.POutput, w io.Writer) error {
	if po.RedeemScript != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.RedeemScriptOutputType), nil, po.RedeemScript,
		)
		if err != nil {
			return err
		}
	}
	if po.WitnessScript != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.WitnessScriptOutputType), nil, po.WitnessScript,
		)
		if err != nil {
			return err
		}
	}

	sort.Sort(psbt.Bip32Sorter(po.Bip32Derivation))
	for _, kd := range po.Bip32Derivation {
		err := serializeKVPairWithType(w,
			uint8(psbt.Bip32DerivationOutputType),
			kd.PubKey,
			psbt.SerializeBIP32Derivation(
				kd.MasterKeyFingerprint,
				kd.Bip32Path,
			),
		)
		if err != nil {
			return err
		}
	}

	if po.TaprootInternalKey != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.TaprootInternalKeyOutputType), nil,
			po.TaprootInternalKey,
		)
		if err != nil {
			return err
		}
	}

	if po.TaprootTapTree != nil {
		err := serializeKVPairWithType(
			w, uint8(psbt.TaprootTapTreeType), nil,
			po.TaprootTapTree,
		)
		if err != nil {
			return err
		}
	}

	sort.Slice(po.TaprootBip32Derivation, func(i, j int) bool {
		return po.TaprootBip32Derivation[i].SortBefore(
			po.TaprootBip32Derivation[j],
		)
	})
	for _, derivation := range po.TaprootBip32Derivation {
		value, err := psbt.SerializeTaprootBip32Derivation(
			derivation,
		)
		if err != nil {
			return err
		}
		err = serializeKVPairWithType(
			w, uint8(psbt.TaprootBip32DerivationOutputType),
			derivation.XOnlyPubKey, value,
		)
		if err != nil {
			return err
		}
	}

	// Unknown is a special case; we don't have a key type, only a key and
	// a value field
	for _, kv := range po.Unknowns {
		err := serializeKVpair(w, kv.Key, kv.Value)
		if err != nil {
			return err
		}
	}

	return nil
}
