package main

/*
 * 切片的内部实现和基础功能:
 * 切片这种数据结构便于使用和管理数据集合
 * 切片围绕动态数组的概念构建, 可按需增长和缩小
 *
 * 内置函数 append 实现切片的动态增长, 可以快速且高效的增长切片，
 * 还可以通过切片再次切片来缩小一个切片的大小
 *
 * 切片的底层内存是在连续块中分配的，所以切片能获得索引, 迭代以及
 * 为垃圾回收优化的好处
 */

// 理解:
// 切片可以使用索引，且非常高效， 即make即用，其内存一定是连续的，
// 切片可以动态增长似乎看起来其底层数据结构是链表，但内置函数
// append 的原理是如果增长的长度超出切片底层数组的容量就会创建
// 一个新的底层数组来存放append后的数据，所以切片的增长是新开辟
// 出一片新的连续的内存，所以切片的底层内存是连续的
// 应该注意到，切片是一个很小的对象，其对底层数组进行了抽象，
// 并提供相关的操作方法，其有三个字段的数据结构，这些数据结构
// 包含了go语言操作底层数组的元数据

// 切片的三个字段:
// 指向底层数组的指针
// 切片可访问的元素个数(长度)
// 切片允许增长到的元素个数(容量)

// 相关操作:
// append 增长切片
// len 获取切片长度
// cap 获取切片容量
// 内置函数 len 和 cap 可以用于处理数组、切片和通道

import (
	"fmt"
	"time"
)

func grammar() {
	//--------------------------------------------------------- 概念语法相关

	// 1, 使用make创建切片
	// slice := make([]string, 5)    // 长度:5,  容量:5
	// slice := make([]string, 3, 5) // 长度:3， 容量:5
	// 底层数据的长度是指定的容量，但是初始化后并不能访问所有的数组元素，
	// 不允许创建容量小于长度的切片

	// 2, 使用切片字面量创建切片
	// 切片之所以被称为切片，是因为创建一个新的切片就是把底层数组切出一部分.
	// slice := []int{10, 20, 30, 40, 50} // 长度:5,  容量:5
	// newSlice := slice[1:3]             // 长度:2,  容量:4
	// newSlice := append(newSlice, 60)   // 长度:3,  容量:4
	//
	//			slice:=[]int{10, 20, 30, 40, 50}
	//			地址指针，长度(5), 容量(5)
	//               |
	//		         |
	//		  _______|
	//       |
	//       |
	//      \|/
	//      [0]		[1]		[2]		[3]			[4]
	//      10		20		30		40---60		50     底层数组
	//				[0]		[1]		[2]		    [3]
	//               /|\
	//				  |
	//                |________________________
	//       ________ |                       |
	//		|                                 |
	//		|                                 |
	//		|                                 |
	//  地址指针，长度(2), 容量(4)            |
	//   newSlice := slice[1:3]               |
	//                                        |
	//										地址指针，长度(3), 容量(4)
	//									newSlice = append(newSlice, 60)
	//
	// slice 和 newSlice 共享同一段底层数组，但通过不同的切片会看到底层数组
	// 不同部分, 因为两个切片共享同一个底层数组，如果一个切片修改了该底层数组
	// 的共享部分，则另一个切片访问到的数据也随之改变
	// 与切片的容量相关联的元素只能用于增长切片，在使用这部分切片元素前，必须
	// 将其合并到切片的长度里
	//
	// append, newSlice 在底层数组中有额外的容量可用，append 操作将可用的元素
	// 合并到切片的长度，并对其进行赋值, 由于和原始的 slice 共享同一个底层数
	// 组，slice 中索引为3的元素也被改动了
	//

	/*
	 *内置函数append会处理增加切片长度时的所有操作细节
	 * append会返回一个包含修改结果的新切片
	 * append总是会增加新切片的长度，而容量有可能会改变，也可能不改变， 取决于
	 * 被操作的切片的可用容量
	 */

	//
	// 如果切片的底层数组没有足够的可用容量，append函数会创建一个新的底层数组,
	// 将被引用的值复制到新数组里，再追加新的值
	// slice := []int{10, 20, 30, 40} // 长度:4, 容量:4
	// 向切片追加一个新元素
	// newSlice := append(slice, 50)  // 长度:5, 容量:8
	// append 后，newSlice 拥有一个全新的底层数组，这个数组的容量是原来的两倍
	//
	//      slice := []int{10, 20, 30, 40}
	//      地址指针，长度(4), 容量(4)
	//       |
	//	     |
	//	     |
	//       |
	//       |
	//      \|/
	//      [0]		[1]		[2]		[3]
	//      10		20		30		40    底层数组
	//
	//
	//		newSlice := append(slice, 50)
	//		地址指针，长度(2), 容量(4)
	//       |
	//	     |
	//	     |
	//       |
	//       |
	//      \|/
	//      [0]		[1]		[2]		[3]		[4]		[5]		[6]		[7]
	//      10		20		30		40		50		0		0		0  底层数组
	//
	//  append 会智能地处理底层数组的容量增长，在切片的容量小于1000个元素时，
	//	总是成倍地增加容量. 若元素个数超过1000，容量的增长因子会设为1.25，也
	//	是每次增加25%的容量.
	//
	// slice := []string{99: ""}
	// 使用空字符串初始化第100个元素
	// 注意: 如果在 [] 运算符里指定了一个值，创建的就是数组而不是切片
	//
	//
	// 一个切片追加到另一个切片
	// s1 := []int{1, 2} 长度:2 容量:2
	// s2 := []int{3, 4} 长度:2 容量:2
	// s3 := []int{5, 6, 7} 长度:2 容量:2
	// slice := append(s1, s2...)  	// 长度: 4, 容量: 4
	// slice := append(s1, s3...)  	// 长度: 5, 容量: 6
	//
	//
	// 3, nil 和空切片
	// var slice []int // 创建 nil 整型切片
	// 在声明时不做任何初始化, 没有分配内存空间
	// slice 的地址指针为 nil, 其长度和容量为0
	// ----在需要描述一个不存在的切片时， nil切片很有用
	//
	// slice := make([]int, 0)
	// 使用make创建空切片
	// 注意：使用make函数创建slice时，至少传入长度参数
	// 空切片在底层数组包含0个元素，也没有分配任何存储空间, 但有地址指针?
	// ----在表示空集合时空切片很有用
	//
	//
	//
	//
	// nil切片和空切片，对其调用内置函数append, len, cap的效果都是一样的
	// todo: nil切片和空切片的区别是什么???
	//---------------------------------------------------------

}

func main() {
	//  分配包含100万个整型值的切片
	slice := make([]int, 1e6)

	// 传递给函数 foo
	start := time.Now()
	slice = foo(slice)
	duration := time.Since(start)
	fmt.Println(duration)
}

func foo(slice []int) []int {
	// ...
	return slice
}

// foo
// 142ns, 对比 ./array.go 中数组在函数间的传递，就可发现传递切片的性能较好
// 由于切片的尺寸很小， 在函数间传递切片成本很低

// 在 64 位架构上，一个切片需要 24 字节的内存: 如下
// 指针字段需要 8 字节（内存按位寻址， 表示一个地址需要64位, 即8字节）
// 长度字段需要 8 字节(?)
// 容量字段需求 8 字节(?)
// 由于与切片关联的数据包含在底层数组里， 不属于切片本身，所有将切片复制到
// 任意函数的时候， 对底层数组的大小都不会有影响，复制时只会复制切片本身，
// 不会涉及到底层数组, 也即在函数间传递数组传递的是要共享的底层数组的相关
// 信息，而不是直接传递整个数组

//
//
// 在创建切片时，可以使用第三个索引来控制新切片的容量，允许限制新切片的容量
// 为底层数组提供了一定的保护，可以更好的控制追加操作.
// slice := []int{10, 20, 30, 40, 50}
// newSlice := slice[2:3:4]  // 长度:1, 容量:2
// 由前所述，内置函数append首先会使用可用容量，一旦没有可用容量，会分配
// 一个新的底层数组.
// 如果两个切片共享一段底层数组，那么对其中一个切片的修改，很可能导致
// 随机且奇怪的问题，对切片内容的修改会影响多个切片，却很难找到原因.
// 如果在创建切片时设置切片的容量和长度一样，就可以强制让新切片的第一个
// append 操作创建新的底层数组，与原有的底层数组分离, 就可以安全的进行
// 后续修改
//
//
// 注意: 使用range迭代切片时，range 创建了每个元素的副本， 而不是直接返回对该
// 元素的引用，迭代返回的变量是一个迭代过程中根据切片依次赋值的新变量，类似于
// range 申请的临时变量来存放每次迭代的数据，所以在for range 中如果需要取每个
// 元素的地址，需要使用切片的索引进行取地址操作.
//
