package dt

import (
	"log"
	"reflect"
	"sync"
	"testing"
)

func NewStackPush(length int32, val ...interface{}) Stack {
	stack := NewStack(length, LockStack)
	for _, v := range val {
		stack.Push(v)
	}
	return stack
}

func getStacks() map[string]Stack {
	return map[string]Stack{
		"zero":       NewStack(0, LockStack),
		"empty":      NewStack(5, LockStack),
		"full":       NewStackPush(1, "full"),
		"remain_one": NewStackPush(5, "one", "two", "three", "four"),
	}
}

func TestStack_Pop(t *testing.T) {
	tests := []struct {
		name string
		want interface{}
	}{
		{"zero", nil},
		{"empty", nil},
		{"full", "full"},
		{"full", nil},
		{"remain_one", "four"},
		{"remain_one", "three"},
	}
	stacks := getStacks()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stacks[tt.name]
			if got := s.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_Push(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{}, true},
		{"empty", args{1}, true},
		{"zero", args{1}, false},
		{"full", args{1}, false},
		{"remain_one", args{"five"}, true},
		{"remain_one", args{"six"}, false},
	}
	stacks := getStacks()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stacks[tt.name]
			if got := s.Push(tt.args.val); got != tt.want {
				t.Errorf("Push() = %v, want %v", got, tt.want)
			}
		})
	}
}

// random test

func TestWrite(t *testing.T) {
	stack := NewStack(20, SliceStack)
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			for x := 0; x < 1e6; x++ {
				ret := stack.Push(i)
				if ret != true {
					t.Error("push err", i)
					continue
				}

				y := stack.Pop()
				if y == nil {
					t.Error("err pop y is nil:", i)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	remain := 0
	for {
		x := stack.Pop()
		if x == nil {
			break
		}
		remain++
	}
	log.Println("remain:", remain)
}

func TestBase(t *testing.T) {
	tests := []struct {
		name  string
		stack STACK
	}{
		{"link", LinkStack},
		{"cas", CasStack},
		{"slice", SliceStack},
		{"lock", LockStack},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stack := NewStack(1000, test.stack)
			for i := 0; i < 100; i++ {
				stack.Push(i)
			}
			for i := 0; i < 50; i++ {
				stack.Pop()
			}

			if rem := remain(stack); rem != 50 {
				t.Error("remain:", rem)
			}
		})
	}
}

func BenchmarkStack(b *testing.B) {
	benchmarks := []struct {
		name  string
		stack STACK
	}{
		{"link", LinkStack},
		{"cas", CasStack},
		{"slice", SliceStack},
		{"lock", LockStack},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			stack := NewStack(1000, bm.stack)
			for i := 0; i < b.N; i++ {
				stack.Push(i)
				stack.Pop()
			}
			if rem := remain(stack); rem != 0 {
				b.Error("remain:", rem)
			}
		})
	}
}

func remain(stack Stack) int {
	i := 0
	for {
		if stack.Pop() == nil {
			break
		}
		i++
	}
	return i
}
