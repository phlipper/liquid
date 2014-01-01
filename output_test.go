package liquid

import (
	"github.com/karlseguin/gspec"
	"github.com/karlseguin/liquid/filters"
	"strconv"
	"testing"
)

func TestOutputHandlesEmptyOutput(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{}}"))
	spec.Expect(output).ToBeNil()
	spec.Expect(err).ToBeNil()
}

func TestOutputHandlesSpaceOnlyOutput(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{   }}"))
	spec.Expect(output).ToBeNil()
	spec.Expect(err).ToBeNil()
}

func TestOutputExtractsASimpleStatic(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{  'over 9000'}}"))
	spec.Expect(string(output.(*StaticOutput).Value)).ToEqual("over 9000")
	spec.Expect(err).ToBeNil()
}

func TestOutputExtractsAComplexStatic(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{'it\\'s over \\9000'}}"))
	spec.Expect(string(output.(*StaticOutput).Value)).ToEqual("it's over \\9000")
	spec.Expect(err).ToBeNil()
}

func TestOutputExtractsAStaticWithAnEndingQuote(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{'it\\''}}"))
	spec.Expect(string(output.(*StaticOutput).Value)).ToEqual("it'")
	spec.Expect(err).ToBeNil()
}

func TestOutputExtractionGivesErrorForUnclosedStatic(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{ 'failure }}"))
	spec.Expect(output).ToBeNil()
	spec.Expect(err.Error()).ToEqual(`Output had an unclosed single quote in "{{ 'failure }}"`)
}

func TestOutputExtractASimpleDynamic(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{ name  }}"))
	assertDynamic(spec, output, "name")
	spec.Expect(err).ToBeNil()
}

func TestOutputExtractANestedDynamic(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{ user.name.first  }}"))
	assertDynamic(spec, output, "user", "name", "first")
	spec.Expect(err).ToBeNil()
}

func TestOutputNoFiltersForStatic(t *testing.T) {
	spec := gspec.New(t)
	output, _ := outputExtractor([]byte("{{'fun'}}"))
	spec.Expect(len(output.(*StaticOutput).Filters)).ToEqual(0)
}

func TestOutputGeneratesErrorOnUnknownFilter(t *testing.T) {
	spec := gspec.New(t)
	_, err := outputExtractor([]byte("{{'fun' | unknown }}"))
	spec.Expect(err.Error()).ToEqual(`Unknown filter "unknown" in "{{'fun' | unknown }}"`)
}

func TestOutputGeneratesErrorOnInvalidParameter(t *testing.T) {
	spec := gspec.New(t)
	_, err := outputExtractor([]byte("{{'fun' | debug: 'missing }}"))
	spec.Expect(err.Error()).ToEqual(`Missing closing quote for parameter in "{{'fun' | debug: 'missing }}"`)
}

func TestStaticOutputWithASingleFilter(t *testing.T) {
	output, _ := outputExtractor([]byte("{{'fun' | debug }}"))
	assertFilters(t, output.(*StaticOutput).Filters, "debug(0)")
}

func TestStaticOutputWithMultipleFilters(t *testing.T) {
	output, _ := outputExtractor([]byte("{{'fun' | debug | debug}}"))
	assertFilters(t, output.(*StaticOutput).Filters, "debug(0)", "debug(1)")
}

func TestStaticOutputWithMultipleFiltersHavingParameters(t *testing.T) {
	spec := gspec.New(t)
	output, err := outputExtractor([]byte("{{'fun' | debug:1,2 | debug:'test' | debug : 'test' , 5}}"))
	spec.Expect(err).ToBeNil()
	assertFilters(t, output.(*StaticOutput).Filters, "debug(0, 1, 2)", "debug(1, test)", "debug(2, test, 5)")
}

// func TestOutputNoFiltersForDynamic(t *testing.T) {
// 	spec := gspec.New(t)
// 	output, _ := outputExtractor([]byte("{{ fun }}"))
// 	spec.Expect(len(output.(*DynamicOutput).Filters)).ToEqual(0)
// }

func assertDynamic(spec *gspec.S, output Token, expected ...string) {
	d := output.(*DynamicOutput)
	spec.Expect(len(d.Fields)).ToEqual(len(expected))
	for index, e := range expected {
		spec.Expect(string(d.Fields[index])).ToEqual(e)
	}
}

func assertFilters(t *testing.T, filters []filters.Filter, expected ...string) {
	spec := gspec.New(t)
	spec.Expect(len(filters)).ToEqual(len(expected))
	for index, filter := range filters {
		actual := string(filter(strconv.Itoa(index)).([]byte))
		spec.Expect(actual).ToEqual(expected[index])
	}
}
