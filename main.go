package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"os"

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
	results := kmeans(imageVectors)

	fmt.Println(results)
}

func kmeans(vectors []imageVector) []imageVector {
	resultVector := []imageVector{}

	for i := 1; i <= 6; i++ {
		tmpVector := imageVector{colorful.Color{rand.Float64(), rand.Float64(), rand.Float64()}, i}
		resultVector = append(resultVector, tmpVector)
	}

	for _, vector := range vectors {
		vector.group = detectGroup(vector, resultVector)
	}

	//TODO resultVectorの再設定

	//TODO imagevectorsの再クラスタリング

	return resultVector
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
