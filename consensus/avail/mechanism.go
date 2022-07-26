package avail

import "fmt"

type MechanismType string

const (
	//
	Sequencer MechanismType = "sequencer"

	//
	Validator MechanismType = "validator"

	//
	WatchTower MechanismType = "watchtower"
)

// mechanismTypes is the map used for easy string -> mechanism MechanismType lookups
var mechanismTypes = map[string]MechanismType{
	"sequencer":  Sequencer,
	"validator":  Validator,
	"watchtower": WatchTower,
}

// String is a helper method for casting a MechanismType to a string representation
func (t MechanismType) String() string {
	return string(t)
}

// MechanismExists helper function designed to check mechanism existence
func MechanismExists(mechanism MechanismType) bool {
	for _, m := range mechanismTypes {
		if mechanism == m {
			return true
		}
	}
	return false
}

// ParseType converts a mechanism string representation to a MechanismType
func ParseType(mechanism string) (MechanismType, error) {
	// Check if the cast is possible
	castType, ok := mechanismTypes[mechanism]
	if !ok {
		return castType, fmt.Errorf("invalid avail mechanism type %s", mechanism)
	}

	return castType, nil
}

// ParseMechanismConfigTypes converts mechanisms string representations to a list of MechanismType
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
