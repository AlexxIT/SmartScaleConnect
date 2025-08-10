package core

import (
	"time"
)

type Weight struct {
	// main data
	Date   time.Time `json:"Date"`
	Weight float32   `json:"Weight"` // kg

	// almost all scales
	BMI       float32 `json:"BMI,omitempty"`
	BodyFat   float32 `json:"BodyFat,omitempty"`   // percent
	BodyWater float32 `json:"BodyWater,omitempty"` // percent
	BoneMass  float32 `json:"BoneMass,omitempty"`  // kg

	// some scales
	MetabolicAge   int     `json:"MetabolicAge,omitempty"`   // years
	MuscleMass     float32 `json:"MuscleMass,omitempty"`     // kg
	PhysiqueRating int     `json:"PhysiqueRating,omitempty"` // 1-9
	ProteinMass    float32 `json:"ProteinMass,omitempty"`    // kg
	VisceralFat    int     `json:"VisceralFat,omitempty"`    // 1-12 or 13-59

	// rare scales
	BasalMetabolism    int     `json:"BasalMetabolism,omitempty"`    // kcal, Basal Metabolic Rate (BMR)
	BodyScore          int     `json:"BodyScore,omitempty"`          // points
	HeartRate          int     `json:"HeartRate,omitempty"`          // beats per minute
	Height             float32 `json:"Height,omitempty"`             // cm
	SkeletalMuscleMass float32 `json:"SkeletalMuscleMass,omitempty"` // kg

	User   string `json:"User,omitempty"`
	Source string `json:"Source,omitempty"`

	// unknown data
	//CaloricIntake int // ? Garmin

	// other
	//WHR float32 // Waist-to-Hip Ratio (WHR)
}

func Equal(w1, w2 *Weight) bool {
	return w1.Weight == w2.Weight &&
		w1.BMI == w2.BMI &&
		w1.BodyFat == w2.BodyFat &&
		w1.BodyWater == w2.BodyWater &&
		w1.BoneMass == w2.BoneMass &&
		w1.MetabolicAge == w2.MetabolicAge &&
		w1.MuscleMass == w2.MuscleMass &&
		w1.PhysiqueRating == w2.PhysiqueRating &&
		w1.ProteinMass == w2.ProteinMass &&
		w1.VisceralFat == w2.VisceralFat &&
		w1.BasalMetabolism == w2.BasalMetabolism &&
		w1.BodyScore == w2.BodyScore &&
		w1.HeartRate == w2.HeartRate &&
		w1.Height == w2.Height &&
		w1.SkeletalMuscleMass == w2.SkeletalMuscleMass
}
