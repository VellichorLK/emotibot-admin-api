package model

import (
	"math"
	"math/rand"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

/*
k-means|| algorithm:
Initialization:
1. C <- sample a point uniformly at random from X
2. sum_cost <- cost(C, X)
3. for O(log(sum_cost)) [ATTENTION: here we pick 5] times do:
	C` <- sample each point in X independently with probability  l * d2(x,C) / phi(C,X)
	C <- C with C`
	end for
4. for x in C, set weight to be the number of points in X closer to x than any other point in C
5. Re-cluster C with weights into k clusters
*/

/*
May need to be optimized:
after initialization, the clustering with k centers
*/

const initializationSteps = 5

func Kmeans(
	rawData []Vector,
	k int,
	threshold int) []ClusteredVector {
	rand.Seed(time.Now().UnixNano())
	data := make([]ClusteredVector, len(rawData))
	for idx, vec := range rawData {
		data[idx].Vector = vec
	}

	intiCentroids := seed(data, k)

	return kmeanspp(data, intiCentroids, threshold)
}

func seed(data []ClusteredVector, k int) []Vector {
	logger.Trace.Println("Generate seed starts.")
	centroids := make([]Vector, k)
	costs := make([]float64, len(data))

	for idx := range costs {
		costs[idx] = math.MaxFloat64
	}

	var tmpCentroids []Vector
	tmpCentroids = append(tmpCentroids, data[rand.Intn(len(data))].Vector)
	for i := 0; i < initializationSteps; i++ {
		var sum float64
		logger.Trace.Printf("Loop [%v] for centroids generation.\n", i+1)
		for idx, vec := range data {
			if newCost, _ := pointCost(vec.Vector, tmpCentroids); newCost < costs[idx] {
				costs[idx] = newCost
			}
			sum += costs[idx]
		}
		logger.Trace.Println("Start to sample centroids.")
		for idx := range data {
			possibility := 2.0 * float64(k) * costs[idx] / sum
			if rand.Float64() < possibility {
				tmpCentroids = append(tmpCentroids, data[idx].Vector)
			}
		}
		logger.Trace.Printf("Centroids size: [%v]\n", len(tmpCentroids))
	}

	if len(tmpCentroids) <= k {
		return tmpCentroids
	}

	weights := make([]int, len(tmpCentroids))
	for _, vec := range data {
		_, index := pointCost(vec.Vector, tmpCentroids)
		weights[index]++
	}
	logger.Trace.Printf("Start to re-sampling into [%v] centroids.\n", k)
	start := time.Now()
	centroids = localKMeans(tmpCentroids, weights, k, 30, len(data))
	logger.Trace.Printf("Re-sampling costs: [%v], centroids size: [%v]\n",
		time.Since(start).Seconds(),
		len(centroids))
	return centroids
}

func localKMeans(points []Vector, weights []int, k, threshold, size int) (centers []Vector) {
	centers = make([]Vector, k)
	var (
		dimension int
		curWeight int
		i         int
		r         int
	)
	dimension = len(points[0])
	r = rand.Intn(size)
	for i = 0; i < len(points) && curWeight <= r; i++ {
		curWeight += weights[i]
	}
	centers[0] = points[i-1]

	costArray := make([]float64, len(points))
	for i = 0; i < len(points); i++ {
		costArray[i] = distance(points[i], centers[0])
	}

	for i = 1; i < k; i++ {
		sum := 0.0
		cumulativeScore := 0.0
		var j int
		for j = 0; j < len(points); j++ {
			sum += costArray[j] * float64(weights[j])
		}
		pic := rand.Float64() * sum
		for j = 0; j < len(points) && cumulativeScore <= pic; j++ {
			cumulativeScore += float64(weights[j]) * costArray[j]
		}

		if j == 0 {
			logger.Warn.Printf("kMeansPlusPlus initialization ran out of distinct points for centers."+
				" Using duplicate point for center k = %v\n", i)
			centers[i] = points[0]
		} else {
			centers[i] = points[j-1]
		}

		for idx := range points {
			update := distance(points[idx], centers[i])
			if costArray[idx] > update {
				costArray[idx] = update
			}
		}
	}

	oldClosest := make([]int, len(points))
	for i := 0; i < len(points); i++ {
		oldClosest[i] = -1
	}
	moved := true
	for iteration := 0; iteration < threshold && moved; iteration++ {
		moved = false
		counts := make([]float64, k)
		sums := make([]Vector, k)
		for i := 0; i < k; i++ {
			sums[i] = make(Vector, dimension)
		}
		for i := 0; i < len(points); i++ {
			p := points[i]
			_, idx := pointCost(p, centers)
			p.mul(float64(weights[i]))
			if sums[idx] == nil {
				sums[idx] = p
			} else {
				sums[idx].add(p)
			}
			counts[idx] += float64(weights[i])
			if idx != oldClosest[i] {
				moved = true
				oldClosest[i] = idx
			}
		}

		for j := 0; j < k; j++ {
			if counts[j] == 0.0 {
				centers[j] = points[rand.Intn(len(points))]
			} else {
				sums[j].mul(1.0 / counts[j])
				centers[j] = sums[j]
			}
		}
	}
	return
}

func pointCost(point Vector, centroids []Vector) (minCost float64, index int) {
	minCost = math.MaxFloat64
	for idx, v := range centroids {
		dis := distance(point, v)
		if minCost > dis {
			minCost = dis
			index = idx
		}
	}
	return
}
