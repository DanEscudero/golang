package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
)

type Position struct {
	x, y int
}

type Dimension struct {
	width, height int
}

type Cell struct {
	position Position
	visited  bool
	free     bool
}

type Line = []Cell

type Maze struct {
	start, finish Position
	dimension     Dimension
	cells         []Line
}

type SolvedMaze struct {
	Maze
	path []Position
}

func sliceIntoBooleanMap[T comparable](s []T) map[T]bool {
	elementMap := make(map[T]bool)

	for _, x := range s {
		elementMap[x] = true
	}

	return elementMap
}

func includes[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func initCellMatrix(dimension Dimension) []Line {
	width, height := dimension.width, dimension.height

	matrix := make([]Line, height)
	for i := 0; i < height; i++ {
		matrix[i] = make(Line, width)
		for j := 0; j < width; j++ {
			matrix[i][j] = Cell{}
		}
	}

	return matrix
}

func isBlack(r, g, b, _ uint32) bool {
	avg := (r + g + b) / 3
	return avg < 127
}

func getImage(path string) image.Image {
	input, _ := os.Open(path)
	defer input.Close()

	img, _, _ := image.Decode(input)

	return img
}

func drawRect(img *image.RGBA, p0 image.Point, p1 image.Point, color color.RGBA) {
	for x := p0.X; x <= p1.X; x++ {
		for y := p0.Y; y <= p1.Y; y++ {
			img.Set(x, y, color)
		}
	}
}

func getImageDimensions(img image.Image) Dimension {
	return Dimension{width: img.Bounds().Max.X, height: img.Bounds().Max.Y}
}

func getCellFromPixel(position Position, px color.Color) *Cell {
	return &Cell{position: position, free: !isBlack(px.RGBA())}
}

func getImageCells(img image.Image, dimension Dimension) []Line {
	cells := initCellMatrix(dimension)

	for i, line := range cells {
		for j := range line {
			position := Position{x: j, y: i}
			cell := getCellFromPixel(position, img.At(j, i))
			cells[i][j] = *cell
		}
	}

	return cells
}

func getFreePosition(line Line) Position {
	for _, cell := range line {
		if cell.free {
			return cell.position
		}
	}

	panic("Cant find free cell!")
}

func imageToMaze(img image.Image) *Maze {
	maze := Maze{}

	dimension := getImageDimensions(img)
	maze.dimension = dimension
	maze.cells = getImageCells(img, dimension)

	firstLine, lastLine := maze.cells[0], maze.cells[dimension.height-1]

	maze.start = getFreePosition(firstLine)
	maze.finish = getFreePosition(lastLine)

	return &maze
}

func newImage(dimension Dimension, resolution int) *image.RGBA {
	width, height := dimension.width*resolution, dimension.height*resolution
	p1 := image.Point{0, 0}
	p2 := image.Point{width, height}

	return image.NewRGBA(image.Rectangle{p1, p2})
}

func getMazeCellColor(cell Cell) color.RGBA {
	if cell.free {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

func getPathCellColor(p float32) color.RGBA {
	x := uint8(50 + 150*p)
	return color.RGBA{R: 255, G: x, B: x, A: 255}
}

func getPointWithResolution(p image.Point, resolution int) (image.Point, image.Point) {
	p0 := image.Point{p.X * resolution, p.Y * resolution}
	p1 := image.Point{(p.X+1)*resolution - 1, (p.Y+1)*resolution - 1}

	return p0, p1
}

func (t *SolvedMaze) toImage(resolution int) *image.RGBA {
	img := newImage(t.dimension, resolution)

	for i, line := range t.cells {
		for j, cell := range line {
			p0, p1 := getPointWithResolution(image.Point{j, i}, resolution)
			drawRect(img, p0, p1, getMazeCellColor(cell))
		}
	}

	for i, cell := range t.path {
		p0, p1 := getPointWithResolution(image.Point{cell.x, cell.y}, resolution)
		drawRect(img, p0, p1, getPathCellColor(float32(i)/float32(len(t.path))))
	}

	return img
}

func stringifyLine(line Line, pathMap map[Position]bool) string {
	s := ""
	for _, cell := range line {
		if pathMap[cell.position] {
			s += "+"
		} else if cell.free {
			s += " "
		} else {
			s += "█"
		}
	}

	s += "\n"

	return s
}

func stringifyMazeStats(m Maze) string {
	s := ""
	s += "Dimensions (w x h): " + fmt.Sprint(m.dimension.width) + " x " + fmt.Sprint(m.dimension.height) + "\n"
	s += "Start Position (x,y): (" + fmt.Sprint(m.start.x) + "," + fmt.Sprint(m.start.y) + ")\n"
	s += "Finish Position (x,y): (" + fmt.Sprint(m.finish.x) + "," + fmt.Sprint(m.finish.y) + ")\n"

	return s
}

func (m Maze) String() string {
	s := stringifyMazeStats(m)

	pathMap := sliceIntoBooleanMap(make([]Position, 0))
	for _, line := range m.cells {
		s += stringifyLine(line, pathMap)
	}

	return s
}

func (m SolvedMaze) String() string {
	s := stringifyMazeStats(m.Maze)

	pathMap := sliceIntoBooleanMap(m.path)

	for _, line := range m.cells {
		s += stringifyLine(line, pathMap)
	}

	return s
}

func isInRange(cells [][]Cell, position Position) bool {
	return position.y >= 0 && position.y < len(cells) && position.x >= 0 && position.x < len(cells[0])
}

func getCell(cells [][]Cell, position *Position) Cell {
	return cells[position.y][position.x]
}

func getUnvisitedNeighbors(cells [][]Cell, position *Position) []Position {
	x, y := position.x, position.y

	positions := []Position{{x: x - 1, y: y}, {x: x, y: y + 1}, {x: x + 1, y: y}, {x: x, y: y - 1}}
	available := make([]Position, 0)

	for _, position := range positions {
		if isInRange(cells, position) && getCell(cells, &position).free && !getCell(cells, &position).visited {
			available = append(available, position)
		}
	}

	return available
}

func dfs(cells [][]Cell, start *Position, end *Position, path []Position) (bool, []Position) {
	if *start == *end {
		return true, append(path, *end)
	}

	cells[start.y][start.x].visited = true
	path = append(path, *start)

	neighbors := getUnvisitedNeighbors(cells, start)

	for _, n := range neighbors {
		t, p := dfs(cells, &n, end, path)

		if t {
			return t, p
		}
	}

	return false, make([]Position, 0)
}

func solveMaze(m *Maze) *SolvedMaze {
	_, path := dfs(m.cells, &m.start, &m.finish, make([]Position, 0))

	solved := SolvedMaze{
		Maze: *m,
		path: path,
	}

	return &solved
}

func main() {
	img := getImage("./examples/input/normal.png")

	maze := imageToMaze(img)

	solvedMaze := solveMaze(maze)
	fmt.Println(solvedMaze)

	outputImage := solvedMaze.toImage(12)

	t := time.Now()
	name := t.Format(time.RFC3339)

	f, _ := os.Create("./output-solved/" + name + ".png")
	png.Encode(f, outputImage)
}
