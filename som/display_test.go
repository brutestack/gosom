package som

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/mat"
)

func TestUMatrixSVG(t *testing.T) {
	assert := assert.New(t)

	const svg = `<h1>Done</h1><svg width="140" height="140"><polygon points="65.000000,54.433757 40.000000,68.867513 15.000000,54.433757 15.000000,25.566243 40.000000,11.132487 65.000000,25.566243 65.000000,54.433757 " style="fill:rgb(255,255,255);stroke:black;stroke-width:1"></polygon><polygon points="90.000000,97.735027 65.000000,112.168784 40.000000,97.735027 40.000000,68.867513 65.000000,54.433757 90.000000,68.867513 90.000000,97.735027 " style="fill:rgb(0,0,0);stroke:black;stroke-width:1"></polygon><polygon points="115.000000,54.433757 90.000000,68.867513 65.000000,54.433757 65.000000,25.566243 90.000000,11.132487 115.000000,25.566243 115.000000,54.433757 " style="fill:rgb(0,0,0);stroke:black;stroke-width:1"></polygon><polygon points="140.000000,97.735027 115.000000,112.168784 90.000000,97.735027 90.000000,68.867513 115.000000,54.433757 140.000000,68.867513 140.000000,97.735027 " style="fill:rgb(255,255,255);stroke:black;stroke-width:1"></polygon></svg>`

	mUnits := mat.NewDense(4, 2, []float64{
		0.0, 0.0,
		0.0, 0.1,
		1.0, 1.0,
		1.0, 1.1,
	})
	coordDims := []int{2, 2}
	uShape := "hexagon"
	title := "Done"
	writer := bytes.NewBufferString("")

	UMatrixSVG(mUnits, coordDims, uShape, title, writer, make(map[int]int))

	assert.Equal(svg, writer.String())
	// make sure there is at least one fully black element
	assert.True(strings.Contains(svg, "rgb(0,0,0)"))
	// make sure there is at least one fully white element
	assert.True(strings.Contains(svg, "rgb(255,255,255)"))
}

func TestUMatrixSVGWithClusters(t *testing.T) {
	assert := assert.New(t)

	const svg = `<h1>Done</h1><svg width="90" height="140"><polygon points="45.000000,45.000000 45.000000,-5.000000 -5.000000,-5.000000 -5.000000,45.000000 45.000000,45.000000 " style="fill:rgb(255,0,0);stroke:black;stroke-width:1"></polygon><text x="17.5" y="27.5" style="fill:rgb(128,128,128);">0</text><polygon points="45.000000,95.000000 45.000000,45.000000 -5.000000,45.000000 -5.000000,95.000000 45.000000,95.000000 " style="fill:rgb(0,255,0);stroke:black;stroke-width:1"></polygon><text x="17.5" y="77.5" style="fill:rgb(128,128,128);">1</text></svg>`

	mUnits := mat.NewDense(2, 2, []float64{
		0.0, 0.0,
		1.0, 1.0,
	})
	coordDims := []int{2, 1}
	uShape := "rectangle"
	title := "Done"
	writer := bytes.NewBufferString("")

	classes := map[int]int{
		0: 0,
		1: 1,
	}
	UMatrixSVG(mUnits, coordDims, uShape, title, writer, classes)

	assert.Equal(svg, writer.String())
	// make sure there is at least one text element
	assert.True(strings.Contains(svg, "<text "))
}
