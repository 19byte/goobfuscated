package goobfuscated

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"
)

type ID uint64

const (
	// MaxInt represents the upper bound.
	// It should be 2^N.
	// It is set by default to the upper bound of an (2^53 - 1)
	// which represents the maximum safe integer(Number.MAX_SAFE_INTEGER) in JavaScript.
	MaxInt = 1<<53 - 1 // 2,147,483,647

	// MillerRabin is used to configure the ProbablyPrime function
	// which is used to verify prime numbers.
	// See: https://golang.org/pkg/math/big/#Int.ProbablyPrime
	MillerRabin = 20
)

var (
	// id obfuscate prime number.
	defaultIDPrime      uint64
	defaultIDModInverse uint64
	defaultIDRandom     uint64

	// urlEncoding & urlEncoding is alias for base64.RawURLEncoding and
	// binary.LittleEndian for brevity and consistency in encoding and decoding.
	urlEncoding  = base64.RawURLEncoding
	littleEndian = binary.LittleEndian
)

// primes for obfuscating the id number.
// Downloaded from: http://primes.utm.edu/lists/small/millions/
var primes = []uint64{
	452977333, 452977381, 452977403, 452977411, 452977429, 452977453, 452977463, 452977507,
	452977517, 452977519, 452977573, 452977583, 452977589, 452977601, 452977607, 452977621,
	452977649, 452977703, 452977711, 452977739, 452977747, 452977769, 452977771, 452977783,
	452977807, 452977843, 452977849, 452977853, 452977873, 452977891, 452977913, 452977927,
	452977969, 452977991, 452977997, 452978003, 452978011, 452978017, 452978023, 452978027,
	452978047, 452978051, 452978059, 452978063, 452978081, 452978137, 452978153, 452978171,
	452978249, 452978257, 452978299, 452978321, 452978381, 452978389, 452978419, 452978429,
	452978431, 452978443, 452978453, 452978473, 452978483, 452978503, 452978507, 452978557,
	452978567, 452978587, 452978593, 452978599, 452978609, 452978611, 452978623, 452978627,
	452978633, 452978663, 452978671, 452978683, 452978699, 452978707, 452978723, 452978737,
	452978749, 452978777, 452978803, 452978833, 452978837, 452978849, 452978861, 452978881,
	452978909, 452978941, 452978987, 452978989, 452979017, 452979031, 452979143, 452979173,
	452979187, 452979199, 452979203, 452979251, 452979281, 452979301, 452979311, 452979341,
	452979343, 452979349, 452979353, 452979377, 452979407, 452979463, 452979479, 452979491,
	452979509, 452979511, 452979517, 452979523, 452979539, 452979577, 452979599, 452979617,
	452979661, 452979689, 452979691, 452979701, 452979713, 452979731, 452979743, 452979763,
	452979811, 452979827, 452979851, 452979853, 452979859, 452979881, 452979907, 452979941,
	452979959, 452979979, 452980001, 452980019, 452980037, 452980043, 452980057, 452980061,
	452980063, 452980093, 452980117, 452980127, 452980133, 452980163, 452980201, 452980211,
	452980243, 452980267, 452980279, 452980291, 452980309, 452980321, 452980343, 452980357,
	452980369, 452980373, 452980379, 452980399, 452980471, 452980483, 452980511, 452980523,
	452980531, 452980537, 452980547, 452980553, 452980597, 452980601, 452980613, 452980621,
	452980631, 452980643, 452980681, 452980687, 452980709, 452980721, 452980777, 452980789,
	452980831, 452980921, 452980939, 452980961, 452980967, 452980981, 452980991, 452980993,
	452980999, 452981017, 452981027, 452981029, 452981041, 452981051, 452981069, 452981077,
	452981117, 452981119, 452981129, 452981149, 452981197, 452981203, 452981213, 452981281,
	452981327, 452981357, 452981369, 452981411, 452981447, 452981453, 452981467, 452981483,
	452981519, 452981531, 452981569, 452981587, 452981597, 452981623, 452981647, 452981657,
	452981663, 452981671, 452981689, 452981693, 452981699, 452981731, 452981741, 452981747,
	452981761, 452981777, 452981783, 452981797, 452981801, 452981819, 452981821, 452981833,
}

func init() {
	rand.NewSource(time.Now().UnixNano())

	// Random a PRIME number from local primes. It must be smaller
	// than 2147483647 (MAX ID).
	defaultIDPrime = primes[rand.Intn(len(primes))]

	// defaultIDPrime must be a valid prime.
	if !big.NewInt(int64(defaultIDPrime)).ProbablyPrime(MillerRabin) {
		accuracy := 1.0 - 1.0/math.Pow(float64(4), float64(MillerRabin))
		panic(fmt.Errorf("prime is not a valid prime. [Accuracy: %f]", accuracy))
	}

	// Calculate the Mod Inverse of the Prime number such that
	// (PRIME * INVERSE) & MAX ID == 1.
	defaultIDModInverse = modInverse(int64(defaultIDPrime))
	// Generate a Pure Random Integer less than 2147483647 (MAX ID).
	defaultIDRandom = randN(int64(MaxInt) - 1)
}

// MarshalJSON satisfies json.Marshaller and transparently obfuscates the value
// using Default prime
func (id *ID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON satisfies json.Marshaller and transparently deobfuscates the
// value using inverse of Default prime
func (id *ID) UnmarshalJSON(b []byte) (err error) {
	var s string
	// json.Unmarshal converts a quoted JSON bytes string literal data into an
	// actual string s. The rules are different than for Go, so cannot
	// use strconv.Unquote.
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	// Resolve the base64 string into an ID, if an error is encountered during
	// parsing, this will have a side effect on the ID, and the ID will be reset to zero.
	*id, err = ParseID(s)
	return err
}

// String returns the obfuscated id in base64 string format and with
// little-endian byte order.
func (id *ID) String() string {
	buf := make([]byte, 8)
	littleEndian.PutUint64(buf, id.obfuscate())
	return urlEncoding.EncodeToString(buf)
}

// ParseID is an inverse operation of ID.String(), returns zero if
// any error occurs during parsing.
func ParseID(s string) (ID, error) {
	switch buf, err := urlEncoding.DecodeString(s); {
	case err != nil:
		return 0, fmt.Errorf("fails to decode id: %w", err)
	case len(buf) != 8: // ID expected to be exactly 8 bytes.
		return 0, errors.New("unexpected id format")
	default:
		return ID(DeObfuscate(littleEndian.Uint64(buf))), nil
	}
}

// obfuscate is used to encode n using Knuth's hashing algorithm.
func (id *ID) obfuscate() uint64 { return Obfuscate(id.Value()) }

// deObfuscate is used to decode n back to the original.
// It will only decode correctly if the prime selectors is consistent
// with what was used to encode n.
func (id *ID) deObfuscate(n uint64) { *id = ID(DeObfuscate(n)) }

// Encode & Decode obfuscate and deObfuscate the ID.
func (id *ID) Encode() uint64  { return id.obfuscate() }
func (id *ID) Decode(n uint64) { id.deObfuscate(n) }

// Value returns the raw integer value.
func (id *ID) Value() uint64 { return uint64(*id) }

// IsZero reports if the id is the zero value.
func (id *ID) IsZero() bool { return *id == 0 }

// Obfuscate is used to encode id using Knuth's hashing algorithm.
func Obfuscate(id uint64) uint64 { return ((id * defaultIDPrime) & MaxInt) ^ defaultIDRandom }

// DeObfuscate is used to decode n back to the original id.
// It will only decode correctly if the prime selectors is consistent
// with what was used to encode n.
func DeObfuscate(n uint64) uint64 { return ((n ^ defaultIDRandom) * defaultIDModInverse) & MaxInt }

// modInverse returns the modular inverse of a given prime number.
// The modular inverse is defined such that
// (PRIME * MODULAR_INVERSE) & (MAX_INT_VALUE) = 1.
//
// See: http://en.wikipedia.org/wiki/Modular_multiplicative_inverse
//
// NOTE: prime is assumed to be a valid prime. If prime is outside the bounds of
// an int64, then the function panics as it can not calculate the mod inverse.
func modInverse(prime int64) uint64 {
	max := big.NewInt(MaxInt + 1)
	return (&big.Int{}).ModInverse(big.NewInt(prime), max).Uint64()
}

// randN returns a cryptographically secure random number
// in the range [1,N].
func randN(N int64) uint64 {
	n, _ := crand.Int(crand.Reader, big.NewInt(N))
	in := n.Uint64() + 1
	return in
}
