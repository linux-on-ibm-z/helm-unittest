package validators

import (
	"fmt"
	"reflect"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// EqualOrLessValidator validate whether the value of Path is less or equal to Value
type EqualOrLessValidator struct {
	Path  string
	Value interface{}
}

func (a EqualOrLessValidator) failInfo(msg string, index int, not bool) []string {
	return splitInfof(
		setFailFormat(not, true, false, false, " to be greater or equal, got"),
		index,
		a.Path,
		msg,
	)
}

// validate performs a validation of a Kubernetes manifest against an expected value.
// It compares the actual value retrieved from the manifest with the expected value,
// ensuring that they are of compatible types and that the actual value is less than or equal to the expected value.
// If successful, it returns true along with a nil error slice. If unsuccessful, it returns false
// along with an error slice containing information about the validation failure.
func (a EqualOrLessValidator) validate(expected, actual interface{}) (bool, []string) {

	expStr := fmt.Sprintf("%v", expected)
	actStr := fmt.Sprintf("%v", actual)

	switch exp := expected.(type) {
	case string:
		if exp <= actual.(string) {
			return true, nil
		}
	case int:
		if exp <= actual.(int) {
			return true, nil
		}
	case float64:
		if exp <= actual.(float64) {
			return true, nil
		}
	default:
		return false, []string{fmt.Sprintf("unsupported type '%T'", expected)}
	}

	return false, []string{fmt.Sprintf("the expected '%s' is not less or equal to the actual '%s'", expStr, actStr)}
}

// Validate implement Validatable
func (a EqualOrLessValidator) Validate(context *ValidateContext) (bool, []string) {
	log.WithField("validator", "le").Debugln("expected content:", a.Value, "path:", a.Path)
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
		if err != nil {
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if len(actual) == 0 {
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown path '%s'", a.Path))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		actType := reflect.TypeOf(actual[0])
		expType := reflect.TypeOf(a.Value)

		if actType != expType {
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("actual '%s' and expected '%s' types do not match", actType, expType))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		result, errors := a.validate(a.Value, actual[0])
		if errors != nil {
			errorMessage := a.failInfo(errors[0], idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		}
		validateSuccess = determineSuccess(idx, validateSuccess, result)
		break
	}

	return validateSuccess, validateErrors
}
