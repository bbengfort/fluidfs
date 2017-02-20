package fluid_test

import (
	"fmt"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {

	Describe("handling", func() {

		It("should correctly update a version", func() {
			alpha := &Version{1, 239102, 239102}
			bravo := &Version{2, 239108, 239108}

			alpha.Update(bravo)
			Ω(alpha.Latest).Should(Equal(bravo.Latest))
			Ω(alpha.Latest).Should(Equal(uint64(239108)))

			alpha = &Version{1, 239102, 239102}
			bravo = &Version{2, 239108, 239108}

			bravo.Update(alpha)
			Ω(alpha.Latest).Should(Equal(bravo.Latest))
			Ω(alpha.Latest).Should(Equal(uint64(239108)))
		})

		It("should correctly generate the next version", func() {
			alpha := &Version{1, 239102, 239108}
			bravo := alpha.Next(uint(3))

			Ω(bravo.PID).Should(Equal(uint(3)))
			Ω(bravo.Scalar).Should(Equal(uint64(239109)))
			Ω(bravo.Latest).Should(Equal(uint64(239109)))
			Ω(alpha.Latest).Should(Equal(uint64(239109)))
		})

		It("should maintain the PID on next", func() {
			alpha := &Version{1, 239102, 239102}
			bravo := alpha.Next(0)

			Ω(bravo.PID).Should(Equal(uint(1)))
			Ω(bravo.Scalar).Should(Equal(uint64(239103)))
			Ω(bravo.Latest).Should(Equal(uint64(239103)))
			Ω(alpha.Latest).Should(Equal(uint64(239103)))
		})

	})

	Describe("comparison", func() {
		It("should determine if two versions are equal", func() {

			alpha := &Version{1, 239102, 239102}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 239102, 239102, true},
				{2, 239102, 239102, false},
				{1, 239104, 239104, false},
				{3, 129323, 129323, false},
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				Ω(alpha.Equal(bravo)).Should(Equal(tt.expect))
			}
		})

		It("should determine if two versions are not equal", func() {

			alpha := &Version{1, 239102, 239102}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 239102, 239102, false},
				{2, 239102, 239102, true},
				{1, 239104, 239104, true},
				{3, 129323, 129323, true},
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				Ω(alpha.NotEqual(bravo)).Should(Equal(tt.expect))
			}
		})

		It("should determine if a version is less than another", func() {

			alpha := &Version{8, 821923, 821923}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 821922, 821922, true},   // PID less | Scalar less
				{2, 821923, 821923, true},   // PID less | Scalar equal
				{3, 821924, 821924, false},  // PID less | Scalar greater
				{8, 821922, 821922, true},   // PID equal | Scalar less
				{8, 821923, 821923, false},  // PID equal | Scalar equal
				{8, 821924, 821924, false},  // PID equal | Scalar greater
				{9, 821922, 821922, true},   // PID greater | Scalar less
				{11, 821923, 821923, false}, // PID greater | Scalar equal
				{31, 821924, 821924, false}, // PID greater | Scalar greater
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				fmts := "%s < %s is not %t"
				Ω(bravo.Less(alpha)).Should(Equal(tt.expect), fmt.Sprintf(fmts, bravo, alpha, tt.expect))

				if bravo.NotEqual(alpha) {
					Ω(alpha.Less(bravo)).Should(Equal(!tt.expect), fmt.Sprintf(fmts, alpha, bravo, !tt.expect))
				} else {
					Ω(alpha.Less(bravo)).Should(Equal(tt.expect), fmt.Sprintf(fmts, alpha, bravo, tt.expect))
				}
			}
		})

		It("should determine if a version is less or equal than another", func() {

			alpha := &Version{8, 821923, 821923}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 821922, 821922, true},   // PID less | Scalar less
				{2, 821923, 821923, true},   // PID less | Scalar equal
				{3, 821924, 821924, false},  // PID less | Scalar greater
				{8, 821922, 821922, true},   // PID equal | Scalar less
				{8, 821923, 821923, true},   // PID equal | Scalar equal
				{8, 821924, 821924, false},  // PID equal | Scalar greater
				{9, 821922, 821922, true},   // PID greater | Scalar less
				{11, 821923, 821923, false}, // PID greater | Scalar equal
				{31, 821924, 821924, false}, // PID greater | Scalar greater
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				fmts := "%s <= %s is not %t"
				Ω(bravo.LessEqual(alpha)).Should(Equal(tt.expect), fmt.Sprintf(fmts, bravo, alpha, tt.expect))

				if bravo.NotEqual(alpha) {
					Ω(alpha.LessEqual(bravo)).Should(Equal(!tt.expect), fmt.Sprintf(fmts, alpha, bravo, !tt.expect))
				} else {
					Ω(alpha.LessEqual(bravo)).Should(Equal(tt.expect), fmt.Sprintf(fmts, alpha, bravo, tt.expect))
				}
			}
		})

		It("should determine if a version is greater than another", func() {

			alpha := &Version{8, 821923, 821923}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 821922, 821922, false}, // PID less | Scalar less
				{2, 821923, 821923, false}, // PID less | Scalar equal
				{3, 821924, 821924, true},  // PID less | Scalar greater
				{8, 821922, 821922, false}, // PID equal | Scalar less
				{8, 821923, 821923, false}, // PID equal | Scalar equal
				{8, 821924, 821924, true},  // PID equal | Scalar greater
				{9, 821922, 821922, false}, // PID greater | Scalar less
				{11, 821923, 821923, true}, // PID greater | Scalar equal
				{31, 821924, 821924, true}, // PID greater | Scalar greater
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				fmts := "%s > %s is not %t"
				Ω(bravo.Greater(alpha)).Should(Equal(tt.expect), fmt.Sprintf(fmts, bravo, alpha, tt.expect))

				if bravo.NotEqual(alpha) {
					Ω(alpha.Greater(bravo)).Should(Equal(!tt.expect), fmt.Sprintf(fmts, alpha, bravo, !tt.expect))
				} else {
					Ω(alpha.Greater(bravo)).Should(Equal(tt.expect), fmt.Sprintf(fmts, alpha, bravo, tt.expect))
				}
			}
		})

		It("should determine if a version is greater or equal than another", func() {

			alpha := &Version{8, 821923, 821923}

			var testCases = []struct {
				pid    uint
				scalar uint64
				latest uint64
				expect bool
			}{
				{1, 821922, 821922, false}, // PID less | Scalar less
				{2, 821923, 821923, false}, // PID less | Scalar equal
				{3, 821924, 821924, true},  // PID less | Scalar greater
				{8, 821922, 821922, false}, // PID equal | Scalar less
				{8, 821923, 821923, true},  // PID equal | Scalar equal
				{8, 821924, 821924, true},  // PID equal | Scalar greater
				{9, 821922, 821922, false}, // PID greater | Scalar less
				{11, 821923, 821923, true}, // PID greater | Scalar equal
				{31, 821924, 821924, true}, // PID greater | Scalar greater
			}

			for _, tt := range testCases {
				bravo := &Version{tt.pid, tt.scalar, tt.latest}
				fmts := "%s >= %s is not %t"
				Ω(bravo.GreaterEqual(alpha)).Should(Equal(tt.expect), fmt.Sprintf(fmts, bravo, alpha, tt.expect))

				if bravo.NotEqual(alpha) {
					Ω(alpha.GreaterEqual(bravo)).Should(Equal(!tt.expect), fmt.Sprintf(fmts, alpha, bravo, !tt.expect))
				} else {
					Ω(alpha.GreaterEqual(bravo)).Should(Equal(tt.expect), fmt.Sprintf(fmts, alpha, bravo, tt.expect))
				}
			}
		})
	})

})
