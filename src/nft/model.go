package nft

import "go.mongodb.org/mongo-driver/bson/primitive"

// define model for collection
type BlockHashValue []byte
type HeightToBlockHash struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value BlockHashValue     `json:"value,omitempty" bson:"value,omitempty"`
}

type InscriptionIDValue []byte
type InscriptionIDToInscriptionEntry struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   InscriptionIDValue `json:"key,omitempty" bson:"key,omitempty"`
	Value InscriptionIDValue `json:"value,omitempty" bson:"value,omitempty"`
}

type SatPointValue []byte
type InscriptionIDToSatPoint struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   InscriptionIDValue `json:"key,omitempty" bson:"key,omitempty"`
	Value SatPointValue      `json:"value,omitempty" bson:"value,omitempty"`
}

type InscriptionNumberToInscriptionID struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value InscriptionIDValue `json:"value,omitempty" bson:"value,omitempty"`
}

type OutPointValue []byte
type OutPointToSatRanges struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   OutPointValue      `json:"key,omitempty" bson:"key,omitempty"`
	Value []int64            `json:"value,omitempty" bson:"value,omitempty"`
}

type OutPointToValue struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   OutPointValue      `json:"key,omitempty" bson:"key,omitempty"`
	Value int64              `json:"value,omitempty" bson:"value,omitempty"`
}

type SatPointToInscriptionID struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   SatPointValue      `json:"key,omitempty" bson:"key,omitempty"`
	Value InscriptionIDValue `json:"value,omitempty" bson:"value,omitempty"`
}

type SatToInscriptionID struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value InscriptionIDValue `json:"value,omitempty" bson:"value,omitempty"`
}

type SatToSatPoint struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value SatPointValue      `json:"value,omitempty" bson:"value,omitempty"`
}

type StatisticToAccount struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value int64              `json:"value,omitempty" bson:"value,omitempty"`
}

type WriteInscriptionStartingBlockCountToTimeStamp struct {
	ID    primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Key   int64              `json:"key,omitempty" bson:"key,omitempty"`
	Value int64              `json:"value,omitempty" bson:"value,omitempty"`
}
