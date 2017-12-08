package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type imageVector struct {
	vector colorful.Color
	group  int
}

func main() {
	filepath := os.Args[1]
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
		return
	}

	imgrect := img.Bounds()

	imageVectors := []imageVector{}

	for h := imgrect.Min.Y; h < imgrect.Max.Y; h++ {
		for w := imgrect.Min.X; w < imgrect.Max.X; w++ {
			c := colorful.MakeColor(img.At(w, h))
			_, S, V := c.Hsv()

			if S > 0.5 && V > 0.5 {
				newVector := imageVector{}
				newVector.vector = c
				imageVectors = append(imageVectors, newVector)
			}
		}
	}
	rand.Seed(time.Now().UnixNano())
	results := kmeans(imageVectors)

	for i, result := range results {
		fmt.Println(i, result.vector.Hex())
	}
	outimage := image.NewRGBA(image.Rect(0, 0, 960, 320))
	for x := 0; x < 960; x++ {
		for y := 0; y < 320; y++ {
			vec := results[int(x/160)].vector
			r, g, b := vec.RGB255()
			c := color.RGBA{r, g, b, 255}
			outimage.Set(x, y, c)
		}
	}

	filename := strings.Split(filepath, ".")
	if len(filename) < 1 {
		return
	}
	newfilename := filename[0] + "-pickupcolor.png"
	f, err = os.OpenFile(newfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	png.Encode(f, outimage)
}

func kmeans(vectors []imageVector) []imageVector {
	resultVectors := []imageVector{}

	//FIXME kmeans++ は初期ベクトルを均等に散らす必要がある
	for i := 1; i <= 6; i++ {
		tmpVector := imageVector{colorful.Color{rand.Float64(), rand.Float64(), rand.Float64()}, i}
		resultVectors = append(resultVectors, tmpVector)
	}

	for i := 0; i < len(vectors); i++ {
		vectors[i].group = detectGroup(vectors[i], resultVectors)
	}

	copyVectors := make([]imageVector, len(resultVectors))
	for {
		copy(copyVectors, resultVectors)
		for i := 0; i < len(resultVectors); i++ {
			resultVectors[i] = resetCenterVector(vectors, resultVectors[i])
		}
		if checkEqual(copyVectors, resultVectors) {
			break
		}

		for i := 0; i < len(vectors); i++ {
			vectors[i].group = detectGroup(vectors[i], resultVectors)
		}
	}

	return resultVectors
}

func detectGroup(vector imageVector, clusters []imageVector) int {
	group := -1
	distance := 1024.0 // 高々 1.0^2 * 3だが一応高めに設定

	for _, cluster := range clusters {
		tmpdistance := (vector.vector.R - cluster.vector.R) * (vector.vector.R - cluster.vector.R)
		tmpdistance += (vector.vector.G - cluster.vector.G) * (vector.vector.G - cluster.vector.G)
		tmpdistance += (vector.vector.B - cluster.vector.B) * (vector.vector.B - cluster.vector.B)

		if distance > tmpdistance {
			distance = tmpdistance
			group = cluster.group
		}
	}
	return group
}

func resetCenterVector(vectors []imageVector, cluster imageVector) imageVector {
	newVector := imageVector{}
	newVector.group = cluster.group
	count := 0.0
	for _, vector := range vectors {
		if cluster.group == vector.group {
			newVector.vector.R += vector.vector.R
			newVector.vector.G += vector.vector.G
			newVector.vector.B += vector.vector.B
			count += 1.0
		}
	}

	if count != 0.0 {
		newVector.vector.R /= count
		newVector.vector.G /= count
		newVector.vector.B /= count
	} else {
		fmt.Println("no members")
	}

	return newVector

}

func checkEqual(prev []imageVector, after []imageVector) bool {
	if len(prev) != len(after) {
		return false
	}
	for i := 0; i < len(prev); i++ {
		if prev[i].vector.R != after[i].vector.R || prev[i].vector.G != after[i].vector.G || prev[i].vector.B != after[i].vector.B || prev[i].group != after[i].group {
			return false
		}
	}
	return true
}
