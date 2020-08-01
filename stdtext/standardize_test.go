package stdtext_test

import (
	"testing"

	"github.com/ourrootsorg/cms-server/stdtext"
	"github.com/stretchr/testify/assert"
)

func TestAsciiFold(t *testing.T) {
	tests := []struct {
		text string
		std  string
	}{
		{
			text: "John Doe",
			std:  "John Doe",
		},
		{
			text: "ÆÐŁØÞßẞæðłþŒœ",
			std:  "AEDLOThssSSaedlthOEoe",
		},
		{
			text: "Ççóòñàáâãä",
			std:  "Ccoonaaaaa",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.std, stdtext.AsciiFold(test.text))
	}
}
