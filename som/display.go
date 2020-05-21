package som

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"

	"gonum.org/v1/gonum/mat"
)

type rowWithDist struct {
	Row  int
	Dist float64
}

type h1 struct {
	XMLName xml.Name `xml:"h1"`
	Title   string   `xml:",innerxml"`
}

type polygon struct {
	XMLName xml.Name `xml:"polygon"`
	Points  []byte   `xml:"points,attr"`
	Style   string   `xml:"style,attr"`
	Id      string   `xml:"id,attr"`
}

type line struct {
	XMLName xml.Name `xml:"line"`
	X1      float64  `xml:"x1,attr"`
	Y1      float64  `xml:"y1,attr"`
	X2      float64  `xml:"x2,attr"`
	Y2      float64  `xml:"y2,attr"`
	Style   string   `xml:"style,attr"`
}

type svgElement struct {
	XMLName xml.Name `xml:"svg"`
	// Width    float64  `xml:"width,attr"`
	// Height   float64  `xml:"height,attr"`
	ViewBox  string `xml:"viewBox,attr"`
	Polygons []interface{}
}

type textElement struct {
	XMLName xml.Name `xml:"text"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
	Style   string   `xml:"style,attr"`
	Text    string   `xml:",innerxml"`
	Id      string   `xml:"id,attr"`
}

var colors = [][]int{{255, 0, 0}, {0, 255, 0}, {0, 0, 255}, {255, 255, 0}, {255, 0, 255}, {0, 255, 255}}

func MakeColors(colorCount int) {
	maxColor := colorCount - 1
	colors = make([][]int, colorCount)
	for i := 0; i < colorCount; i++ {
		colors[i] = []int{
			255 * (maxColor - i) / maxColor,
			255 * i / maxColor,
			0,
		}
	}
}

// UMatrixSVG creates an SVG representation of the U-Matrix of the given codebook.
// It accepts the following parameters:
// codebook - the codebook we're displaying the U-Matrix for
// dims     - the dimensions of the map grid
// uShape   - the shape of the map grid
// title    - the title of the output SVG
// writer   - the io.Writter to write the output SVG to.
// classes  - if the classes are known (i.e. these are test data) they can be displayed providing the information in this map.
// The map is: codebook vector row -> class number. When classes are not known (i.e. running with real data), just provide an empty map
func UMatrixSVG(codebook *mat.Dense, dims []int, uShape, title string, writer io.Writer, classes map[int]int) error {
	xmlEncoder := xml.NewEncoder(writer)
	// array to hold the xml elements
	elems := []interface{}{}
	if title != "" {
		elems = append(elems, h1{Title: title})
	}

	rows, _ := codebook.Dims()
	distMat, err := DistanceMx("euclidean", codebook)
	if err != nil {
		return err
	}
	coords, err := GridCoords(uShape, dims)
	if err != nil {
		return err
	}
	coordsDistMat, err := DistanceMx("euclidean", coords)
	if err != nil {
		return err
	}

	umatrix := make([]float64, rows)
	maxDistance := -math.MaxFloat64
	minDistance := math.MaxFloat64
	for row := 0; row < rows; row++ {
		avgDistance := 0.0
		// this is a rough approximation of the notion of neighbor grid coords
		allRowsInRadius := allRowsInRadius(row, math.Sqrt2*1.01, coordsDistMat)
		for _, rwd := range allRowsInRadius {
			if rwd.Dist > 0.0 {
				avgDistance += distMat.At(row, rwd.Row)
			}
		}
		avgDistance /= float64(len(allRowsInRadius) - 1)
		umatrix[row] = avgDistance
		if avgDistance > maxDistance {
			maxDistance = avgDistance
		}
		if avgDistance < minDistance {
			minDistance = avgDistance
		}
	}
	deltaDistance := maxDistance - minDistance

	// function to scale the coord grid to something visible
	const MUL = 50.0
	const OFF = 20.0
	scale := func(x float64) float64 { return MUL*x + OFF }

	viewBox := fmt.Sprintf("0 0 %f %f", float64(dims[1])*MUL+2*OFF, float64(dims[0])*MUL+2*OFF)
	svgElem := svgElement{
		ViewBox:  viewBox,
		Polygons: make([]interface{}, rows*2+8),
	}
	for row := 0; row < rows; row++ {
		coord := coords.RowView(row)
		var colorMask []int
		classID, classFound := classes[row]
		// if no class information, just use shades of gray
		if !classFound || classID == -1 {
			colorMask = []int{255, 255, 255}
		} else {
			colorMask = colors[classes[row]%len(colors)]
		}

		colorMul := 1.0
		if deltaDistance > 0 {
			colorMul -= (umatrix[row] - minDistance) / deltaDistance
		}
		r := int(colorMul * float64(colorMask[0]))
		g := int(colorMul * float64(colorMask[1]))
		b := int(colorMul * float64(colorMask[2]))
		polygonCoords := ""
		x := scale(coord.At(0, 0))
		y := scale(coord.At(1, 0))
		// hexagon has a different yOffset
		var textOffsetX, textOffsetY float64
		switch uShape {
		case "hexagon":
			{
				xOffset := 0.5 * MUL
				yBigOffset := math.Tan(math.Pi/6.0) * MUL
				ySmallOffset := yBigOffset / 2.0
				// draw a hexagon around the current coord
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x+xOffset, OFF+y+ySmallOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x, OFF+y+yBigOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x-xOffset, OFF+y+ySmallOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x-xOffset, OFF+y-ySmallOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x, OFF+y-yBigOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x+xOffset, OFF+y-ySmallOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", OFF+x+xOffset, OFF+y+ySmallOffset)
				textOffsetX = 1.25 * OFF
				textOffsetY = 0.75 * OFF
			}
		default:
			{
				xOffset := 0.5 * MUL
				yOffset := 0.5 * MUL
				// draw a box around the current coord
				polygonCoords += fmt.Sprintf("%f,%f ", x+xOffset, y+yOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", x+xOffset, y-yOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", x-xOffset, y-yOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", x-xOffset, y+yOffset)
				polygonCoords += fmt.Sprintf("%f,%f ", x+xOffset, y+yOffset)

				textOffsetX = 0.5 * OFF
				textOffsetY = -0.25 * OFF
			}
		}

		svgElem.Polygons[row*2] = polygon{
			Points: []byte(polygonCoords),
			Style:  fmt.Sprintf("fill:rgb(%d,%d,%d);stroke:black;stroke-width:1", r, g, b),
			Id:     fmt.Sprintf("poly%d", row),
		}

		// print class number
		if classFound {
			svgElem.Polygons[row*2+1] = textElement{
				X:     textOffsetX + x - 0.25*MUL,
				Y:     textOffsetY + y + 0.25*MUL,
				Text:  fmt.Sprintf("%d", classes[row]),
				Style: fmt.Sprintf("fill:rgb(%d,%d,%d);", (r+128)%255, (g+128)%255, (b+128)%255),
				Id:    fmt.Sprintf("text%d", row),
			}
		}
	}

	yStep := scale(coords.RowView(1).At(1, 0)) - scale(coords.RowView(0).At(1, 0))
	xStep := scale(coords.RowView(dims[0]).At(0, 0)) - scale(coords.RowView(0).At(0, 0))
	lineStyle := "stroke:black;stroke-width:3"
	coord := coords.RowView(0)
	x1 := scale(coord.At(0, 0))
	y1 := scale(coord.At(1, 0))
	coord = coords.RowView(rows - 1)
	x2 := scale(coord.At(0, 0))
	y2 := scale(coord.At(1, 0))

	svgElem.Polygons[len(svgElem.Polygons)-1] = line{
		X1:    OFF + x1,
		Y1:    OFF + y1,
		X2:    OFF + x2,
		Y2:    OFF + y2,
		Style: lineStyle,
	}

	coord = coords.RowView(dims[0] - 1)
	x1 = scale(coord.At(0, 0))
	y1 = scale(coord.At(1, 0))
	coord = coords.RowView(rows - dims[0])
	x2 = scale(coord.At(0, 0))
	y2 = scale(coord.At(1, 0))

	svgElem.Polygons[len(svgElem.Polygons)-2] = line{
		X1:    OFF + x1,
		Y1:    OFF + y1,
		X2:    OFF + x2,
		Y2:    OFF + y2,
		Style: lineStyle,
	}

	coord = coords.RowView(dims[0]/2 - 1)
	x1 = scale(coord.At(0, 0))
	y1 = scale(coord.At(1, 0)) + yStep/2
	coord = coords.RowView(rows - dims[0]/2 - 1)
	x2 = scale(coord.At(0, 0))
	y2 = scale(coord.At(1, 0)) + yStep/2

	aLine := line{
		X1:    OFF + x1,
		Y1:    OFF + y1,
		X2:    OFF + x2,
		Y2:    OFF + y2,
		Style: lineStyle,
	}
	svgElem.Polygons[len(svgElem.Polygons)-3] = aLine

	coord = coords.RowView(0)
	x1 = scale(coord.At(0, 0)) + xStep*float64(dims[1]-1)/2
	y1 = scale(coord.At(1, 0))
	x2 = x1
	coord = coords.RowView(dims[0] - 1)
	y2 = scale(coord.At(1, 0))

	bLine := line{
		X1:    OFF + x1,
		Y1:    OFF + y1,
		X2:    OFF + x2,
		Y2:    OFF + y2,
		Style: lineStyle,
	}
	svgElem.Polygons[len(svgElem.Polygons)-4] = bLine

	svgElem.Polygons[len(svgElem.Polygons)-5] = line{
		X1:    aLine.X1,
		Y1:    aLine.Y1,
		X2:    bLine.X1,
		Y2:    bLine.Y1,
		Style: lineStyle,
	}
	svgElem.Polygons[len(svgElem.Polygons)-6] = line{
		X1:    bLine.X1,
		Y1:    bLine.Y1,
		X2:    aLine.X2,
		Y2:    aLine.Y2,
		Style: lineStyle,
	}
	svgElem.Polygons[len(svgElem.Polygons)-7] = line{
		X1:    aLine.X2,
		Y1:    aLine.Y2,
		X2:    bLine.X2,
		Y2:    bLine.Y2,
		Style: lineStyle,
	}
	svgElem.Polygons[len(svgElem.Polygons)-8] = line{
		X1:    bLine.X2,
		Y1:    bLine.Y2,
		X2:    aLine.X1,
		Y2:    aLine.Y1,
		Style: lineStyle,
	}

	elems = append(elems, svgElem)

	xmlEncoder.Encode(elems)
	xmlEncoder.Flush()

	return nil
}

func allRowsInRadius(selectedRow int, radius float64, distMatrix *mat.Dense) []rowWithDist {
	rowsInRadius := []rowWithDist{}
	for i, dist := range distMatrix.RawRowView(selectedRow) {
		if dist < radius {
			rowsInRadius = append(rowsInRadius, rowWithDist{Row: i, Dist: dist})
		}
	}
	return rowsInRadius
}
