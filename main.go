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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type imageVector struct {
	vector colorful.Color
	group  int
}

type imageVectors []imageVector

func (v imageVectors) Len() int {
	return len(v)
}

func (v imageVectors) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v imageVectors) Less(i, j int) bool {
	H1, _, _ := v[i].vector.Hcl()
	H2, _, _ := v[j].vector.Hcl()

	return H1 < H2
}

type kmeansVector struct {
	vector   imageVectors
	distance float64
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

	imageVectors := imageVectors{}

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

	// クラスタ数は仮置き
	clusterSize := 6

	if len(os.Args) > 2 {
		argSize, err := strconv.Atoi(os.Args[2])
		if err == nil {
			clusterSize = argSize
		}
	}

	results, distance := kmeans(imageVectors, clusterSize)

	fmt.Println("distance: ", distance)

	sort.Sort(results)

	for i, result := range results {
		fmt.Println(i, result.vector.Hex())
	}

	imageWidth := 160 * clusterSize
	outimage := image.NewRGBA(image.Rect(0, 0, imageWidth, 320))
	for x := 0; x < imageWidth; x++ {
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
	newfilename := filename[0] + "-pickupcolor" + strconv.Itoa(clusterSize) + ".png"
	f, err = os.OpenFile(newfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	png.Encode(f, outimage)
}

func kmeans(vectors imageVectors, size int) (imageVectors, float64) {
	loopCount := 16
	distance := float64(len(vectors))
	resultVector := make(imageVectors, size)

	vecchan := make(chan *kmeansVector, loopCount)

	wg := &sync.WaitGroup{}

	for count := 0; count < loopCount; count++ {
		wg.Add(1)
		go execKmeans(vectors, size, wg, vecchan)
	}
	wg.Wait()
	close(vecchan)

	for vec := range vecchan {
		if distance > vec.distance {
			copy(resultVector, vec.vector)
			distance = vec.distance
		}
	}
	return resultVector, distance
}

func execKmeans(vectors imageVectors, size int, wg *sync.WaitGroup, ch chan *kmeansVector) {
	resultVectors := initVector(vectors, size)

	for i := 0; i < len(vectors); i++ {
		vectors[i].group = detectGroup(vectors[i], resultVectors)
	}

	copyVectors := make(imageVectors, len(resultVectors))
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

	distance := calcClusterDistance(vectors, resultVectors)

	returnVec := &kmeansVector{
		vector:   resultVectors,
		distance: distance,
	}
	ch <- returnVec
	wg.Done()
}

func calcDistance(vector, cluster imageVector) float64 {
	distance := (vector.vector.R - cluster.vector.R) * (vector.vector.R - cluster.vector.R)
	distance += (vector.vector.G - cluster.vector.G) * (vector.vector.G - cluster.vector.G)
	distance += (vector.vector.B - cluster.vector.B) * (vector.vector.B - cluster.vector.B)

	return distance
}

func calcClusterDistance(vectors, clusters imageVectors) float64 {
	distance := 0.0
	for _, vector := range vectors {
		for _, cluster := range clusters {
			if cluster.group == vector.group {
				distance += calcDistance(vector, cluster)
				break
			}
		}
	}
	return distance
}

//FIXME kmeans++ は初期ベクトルを均等に散らす必要がある
func initVector(vectors imageVectors, size int) imageVectors {
	resultVectors := make(imageVectors, size)

	tmpVectors := make(imageVectors, size)

	for i := 0; i < size; i++ {
		tmpVector := imageVector{colorful.Color{rand.Float64(), rand.Float64(), rand.Float64()}, i + 1}
		tmpVectors[i] = tmpVector
	}

	copy(resultVectors, tmpVectors)

	return resultVectors
}

func detectGroup(vector imageVector, clusters imageVectors) int {
	group := -1
	distance := 1024.0 // 高々 1.0^2 * 3だが一応高めに設定

	for _, cluster := range clusters {
		tmpdistance := calcDistance(vector, cluster)
		if distance > tmpdistance {
			distance = tmpdistance
			group = cluster.group
		}
	}
	return group
}

func resetCenterVector(vectors imageVectors, cluster imageVector) imageVector {
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
		//debug
		//fmt.Println("no members")
	}

	return newVector

}

func checkEqual(prev, after imageVectors) bool {
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
