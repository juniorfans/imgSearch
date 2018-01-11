package kmeans

import (
"math"
"math/rand"
)

type Point struct{
	Entry[] float64
}

type Centroid struct{
	Center Point
	Points []Point
}

func (p_1 Point) DistanceTo(p_2 Point) float64{
	sum := float64(0)
	for e:=0;e<len(p_1.Entry);e++{
		sum += math.Pow((p_1.Entry[e] - p_2.Entry[e]),2)
	}
	return math.Sqrt(sum)
}

/**
	计算当前聚类的质心：使用平均数
	返回旧的质心到新的质心的距离
 */
func (c_1 *Centroid) reCenter() float64{
	new_Centroid := make([]float64,len(c_1.Center.Entry))
	for _, e := range c_1.Points{
		for r:=0;r<len(new_Centroid);r++{
			new_Centroid[r] += e.Entry[r]
		}
	}
	for r:=0;r<len(new_Centroid);r++{
		new_Centroid[r] /= float64(len(c_1.Points))
	}
	old_center := c_1.Center
	c_1.Center = Point{new_Centroid}
	return old_center.DistanceTo(c_1.Center)
}

func KMEANS(data []Point, initCenters []Centroid, k int, DELTA float64) (Centroids []Centroid){

	Centroids = initCenters

	if len(Centroids) > k{
		Centroids = Centroids[0:k]
	}

	//若不够 k 个质心则其它随机指定
	for i:=len(Centroids); i<k; i++{
		Centroids = append(Centroids,Centroid{Center: data[rand.Intn(len(data))]})
	}

	converged := false
	for !converged {
		//以所有数据为对象，计算各自最近的中心，并执行归类
		for i:= range data{
			min_distance := math.MaxFloat64
			z := 0
			for v,e:=range Centroids {
				distance := data[i].DistanceTo(e.Center)
				if distance < min_distance{
					min_distance = distance;
					z = v
				}
			}
			Centroids[z].Points = append(Centroids[z].Points, data[i])
		}

		//计算各个聚类旧质心到新质心的偏移距离的最大值，如果这个最大偏移值比 delta 小则视聚类收敛完成，可以结束。反之需要重新聚类
		max_delta := -math.MaxFloat64
		for i:= range Centroids {
			var movement float64 = 0.0
			if len(Centroids[i].Points) != 0{
				//当前聚类只有一个质心，没有其它附属点，所以质心不会变化
				movement = Centroids[i].reCenter()
			}

			if movement > max_delta{
				max_delta = movement
			}
		}
		//完成收敛
		if DELTA >= max_delta{
			converged = true
			return
		}

		//由当前计算的多个中心，再去聚类
		for i:= range Centroids {
			Centroids[i].Points = nil
		}
	}
	return Centroids
}
