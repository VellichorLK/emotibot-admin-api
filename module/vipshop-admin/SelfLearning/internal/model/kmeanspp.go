package model

import (
	"math"
	"math/rand"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

// Vector: Data Abstraction for an N-dimensional
// vector
type Vector []float64

// Abstracts the Vector with a cluster number
// Update and computation becomes more efficient
type ClusteredVector struct {
	ClusterNumber int
	Distance      float64
	Vector
}

// K-Means Algorithm with smart seeds
// as known as K-Means ++
func KmeansPP(
	rawData []Vector,
	k int,
	threshold int) []ClusteredVector {
	data := make([]ClusteredVector, len(rawData))
	for idx, vec := range rawData {
		data[idx].Vector = vec
	}

	initCentroids := seedpp(data, k)

	return kmeanspp(data, initCentroids, threshold)
}

func KmeansLabels(
	rawData []Vector,
	k int,
	threshold int) []int {
	clusteredData := Kmeans(rawData, k, threshold)
	labels := make([]int, len(clusteredData))
	for i, j := range clusteredData {
		labels[i] = j.ClusterNumber
	}
	return labels
}

func mean(vectors []Vector) (sum Vector) {
	sum = vectors[0]
	for i := 1; i < len(vectors); i++ {
		sum.add(vectors[i])
	}
	for i := range sum {
		sum[i] /= float64(len(vectors))
	}
	return sum
}

// Distance Function: To compute the distance between vectors
func distance(vector0, vector1 []float64) float64 {
	sum := 0.0
	for idx := range vector0 {
		dist := vector0[idx] - vector1[idx]
		sum += dist * dist
	}
	return math.Sqrt(sum)
}

// Summation of two vectors
func (vector Vector) add(otherVector Vector) {
	for i, j := range otherVector {
		vector[i] += j
	}
}

// Multiplication of a vector with a scalar
func (vector Vector) mul(scalar float64) {
	for i := range vector {
		vector[i] *= scalar
	}
}

/*
	@desc : find nearest centroid of $centroids to $clusteredVector
	@return : (nearest centroid, distance)
*/
func near(
	clusteredVector ClusteredVector,
	centroids []Vector) (int, float64) {
	indexOfCluster := 0
	minSquaredDistance := math.MaxFloat64
	for i := 0; i < len(centroids); i++ {
		squaredDistance := distance(clusteredVector.Vector, centroids[i])
		if squaredDistance < minSquaredDistance {
			minSquaredDistance = squaredDistance
			indexOfCluster = i
		}
	}
	return indexOfCluster, math.Sqrt(minSquaredDistance)
}

// Instead of initializing randomly the seeds, make a sound decision of initializing
func seedpp(data []ClusteredVector, k int) []Vector {

	centroids := make([]Vector, k)
	if len(data) == 0 {
		util.LogError.Println("The seed generation fails, caused by: [data size is 0]")
		return centroids
	}

	util.LogTrace.Println("Generate seed starts.")
	centroids[0] = data[rand.Intn(len(data))].Vector
	util.LogTrace.Printf("Generate seed[0] ends. Seed[0]: %v\n",
		centroids[0])
	distPow2 := make([]float64, len(data))

	for i := 1; i < k; i++ {
		var sum float64
		for idx, clusteredVector := range data {
			_, minDist := near(clusteredVector, centroids[:i])
			distPow2[idx] = minDist * minDist
			sum += distPow2[idx]
		}

		target := rand.Float64() * sum
		j := 0
		for sum = distPow2[0]; sum < target; sum += distPow2[j] {
			j++
		}
		centroids[i] = data[j].Vector
	}
	/*
		util.LogTrace.Printf("Generate seed ends, size:[%v], mean[%v]\n",
			len(centroids),
			mean(centroids))
	*/
	return centroids
}

// K-Means Algorithm
func kmeanspp(
	data []ClusteredVector,
	centroids []Vector,
	threshold int) []ClusteredVector {
	counter := 0
	util.LogTrace.Println("Start clustering with k-means")
	for idx, clusteredVector := range data {
		closestCluster, _ := near(clusteredVector, centroids)
		data[idx].ClusterNumber = closestCluster
	}
	util.LogTrace.Println("Initialize the clusters finished")
	sizeCentroids := make([]int, len(centroids))
	dimension := len(data[0].Vector)
	for {
		/*
			calc new centroids
		*/

		util.LogTrace.Printf("Start the [%v]th loop\n", counter)
		for i := range centroids {
			centroids[i] = make(Vector, dimension)
			sizeCentroids[i] = 0
		}

		for _, clusteredVector := range data {
			centroids[clusteredVector.ClusterNumber].add(clusteredVector.Vector)
			sizeCentroids[clusteredVector.ClusterNumber]++
		}

		for i := range centroids {
			centroids[i].mul(1 / float64(sizeCentroids[i]))
		}
		util.LogTrace.Printf("In loop [%v], update the centroids finished\n", counter)
		/*
			find centroid for each node
		*/
		var changes int
		for i, clusteredVector := range data {
			closestCluster, distance := near(clusteredVector, centroids)
			if closestCluster != clusteredVector.ClusterNumber {
				changes++
			}

			data[i].ClusterNumber = closestCluster
			data[i].Distance = distance
		}
		util.LogTrace.Printf("In loop [%v], clustering finished. With [%v] changes.\n",
			counter,
			changes)
		counter++
		if changes == 0 || counter > threshold {
			util.LogTrace.Println("Finish clustering!!")
			return data
		}
	}
}
