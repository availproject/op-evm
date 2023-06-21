package avail

import (
	"fmt"
	"strings"
)

// MechanismType represents the type of a mechanism in the settlement layer system. It is used to categorize and manipulate
// mechanisms based on their types.
type MechanismType string

// MechanismType constants for the possible types of mechanisms.
const (
	BootstrapSequencer MechanismType = "bootstrap-sequencer"
	Sequencer          MechanismType = "sequencer"
	WatchTower         MechanismType = "watchtower"
)

// mechanismTypes is a map used to easily convert a string into its corresponding MechanismType.
var mechanismTypes = map[string]MechanismType{
	"bootstrap-sequencer": BootstrapSequencer,
	"sequencer":           Sequencer,
	"watchtower":          WatchTower,
}

// String is a method for representing a MechanismType as a string.
// It helps in casting a MechanismType to its string representation.
func (t MechanismType) String() string {
	return string(t)
}

// LogString is a method for representing a MechanismType as a string in a log-friendly format.
// It replaces dashes ("-") in the MechanismType with underscores ("_"), for better compatibility with logging systems.
func (t MechanismType) LogString() string {
	return strings.Replace(string(t), "-", "_", -1)
}

// MechanismExists is a helper function designed to check if a given MechanismType exists in the list of mechanism types.
// It returns true if the MechanismType exists and false otherwise.
func MechanismExists(mechanism MechanismType) bool {
	for _, m := range mechanismTypes {
		if mechanism == m {
			return true
		}
	}
	return false
}

// ParseType is a function that converts a string into a MechanismType.
// It checks if the conversion is possible and returns the corresponding MechanismType if it is,
// or an error if the string does not correspond to a known MechanismType.
func ParseType(mechanism string) (MechanismType, error) {
	// Check if the cast is possible
	castType, ok := mechanismTypes[mechanism]
	if !ok {
		return castType, fmt.Errorf("invalid avail mechanism type %s", mechanism)
	}

	return castType, nil
}

// ParseMechanismConfigTypes is a function that converts a list of string representations of mechanism types
// into a slice of MechanismType. It calls ParseType for each string in the list, and returns an error if any string
// does not correspond to a known MechanismType.
func ParseMechanismConfigTypes(mechanisms interface{}) ([]MechanismType, error) {
	mi := mechanisms.([]interface{})
	var toReturn []MechanismType
	for _, i := range mi {
		m, err := ParseType(i.(string))
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, m)
	}

	return toReturn, nil
}
