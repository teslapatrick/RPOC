package minerList

import (
	"encoding/hex"
	"github.com/teslapatrick/RPOC/common"
	"github.com/teslapatrick/RPOC/core/state"
	"github.com/teslapatrick/RPOC/log"
	"github.com/teslapatrick/RPOC/rlp"
	//"github.com/teslapatrick/RPOC/crypto/sha3"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sort"
)

var MinerListContractAddress = common.HexToAddress("0x0665ae1f13f142ad585d32b101c98f531b78c80e")
var KeyMinerLen = "0000000000000000000000000000000000000000000000000000000000000002"
var SelectMod = float64(100) / 100
var EpochTime = 10


type MinerList struct {
	isRegistered map[common.Address]bool
	//honesty      map[common.Address]int
	minerList    []common.Address
	epoch        int32
	selected     common.Address
}

type pair struct {
	key common.Address
	value int
}

type Pair []pair

func NewMinerList() *MinerList {
	ml := &MinerList{
		isRegistered: make(map[common.Address]bool),
		//honesty: make(map[common.Address]int),
	}
	return ml
}

func (ml *MinerList) IsMiner(acc common.Address) bool {
	return ml.isRegistered[acc]
}

func (ml *MinerList) UpdateMinerListSnap(state *state.StateDB) {
	// get miner list in contract storage
	if len(ml.minerList) == 0 {
		return
	}
	minerList := ml.GetMinerList(state)

	// del reg map
	// TODO: find a nice way
	ml.isRegistered = nil
	ml.isRegistered = make(map[common.Address]bool)

	// store miner in the list
	for _, m := range minerList {
		//log.Info("==================>", "m:", m)
		ml.isRegistered[m] = true
	}
}

func MinerLen(state *state.StateDB) *big.Int {
	return state.GetState(MinerListContractAddress, common.HexToHash(KeyMinerLen)).Big()
}

func (ml *MinerList) GetMinerList(state *state.StateDB) []common.Address {
	minerList := make([]common.Address, 0)
	// del miner list map
	ml.minerList = ml.minerList[0:0]
	// get miner list len
	MinerLen := MinerLen(state)
	 if MinerLen == big.NewInt(0) {
	 	return []common.Address{}
	 }
	// do get miner list
	for i:=int64(0); i<MinerLen.Int64(); i++ {
		HashMinerLen := CalculateStateDbIndex(KeyMinerLen, "")
		m := common.BytesToAddress(state.GetState(MinerListContractAddress, common.HexToHash(IncreaseHexByNum(HashMinerLen, i))).Bytes())
		//fmt.Println(">>>>>>>>>>>>>>>getMinerList", m.String())
		ml.minerList = append(ml.minerList, m)
		minerList    = append(minerList, m)
	}
	return minerList
}

func (p Pair) Len() int { return len(p) }
func (p Pair) Less(i, j int) bool { return p[i].value < p[j].value }
func (p Pair) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (ml *MinerList) SortMinerList(honesty map[common.Address]int, signer common.Address) []common.Address {
	miners := ml.minerList

	// del parent signer
	for i:=int(0);i<len(miners);i++  {
		if miners[i] == signer {
			miners = append(miners[:i], miners[i+1:]...)
			delete(honesty, signer)
		}
	}

	// length of minerList
	l := len(miners)
	if l == 0 {
		return []common.Address{}
	}
	// init
	s := make(Pair, l)
	sorted := make([]common.Address, 0)

	// prepare sort data
	for i, m := range miners {
		s[i] = pair{m, honesty[m]}
	}
	// do sort
	// sort.Reverse(interface) 👇
	sort.Sort(sort.Reverse(s))
	// return sorted miner list
	for _, v := range s {
		sorted = append(sorted, v.key)
	}
	return sorted
}

// rand a miner from current miner list
func (ml *MinerList) SelectMiner(preHash common.Hash, epoch int64, honesty map[common.Address]int, parentSigner common.Address) common.Address {
	/*sorted := ml.SortMinerList(honesty, parentSigner)
	//gen rand seed
	randSeed := float64(len(sorted)) * SelectMod
	if randSeed == 0 {
		return common.Address{}
	}
	// cycle
	tempHash := preHash
	var h common.Hash
	for i:=int64(0); i<=epoch; i++ {
		// rlp hash
		h = rlpHash(tempHash)
		tempHash = h
	}

	// selected
	rlpHashBig := h.Big()
	selectedIndex := big.NewInt(0)
	selectedIndex.Mod(rlpHashBig, big.NewInt(int64(randSeed)))
	ml.selected = sorted[selectedIndex.Int64()]
	log.Info("================>", "selected", ml.selected)
	return ml.selected*/

	// new turn
	randSeed := float64(0)
	sorted := []common.Address{}

	if epoch == 0 {
		sorted = ml.SortMinerList(honesty, parentSigner)
		randSeed = float64(len(sorted)) * SelectMod
	} else {
		sorted = ml.SortMinerList(honesty, ml.selected)
		randSeed = float64(len(sorted)) * SelectMod
	}

	if randSeed == 0 {
		log.Error("miner list len is zero.")
		return common.Address{}
	}

	// cycle
	tempHash := preHash
	var h common.Hash
	for i:=int64(0); i<=epoch; i++ {
		// rlp hash
		h = rlpHash(tempHash)
		tempHash = h
	}

	// selected
	rlpHashBig := h.Big()
	selectedIndex := big.NewInt(0)
	selectedIndex.Mod(rlpHashBig, big.NewInt(int64(randSeed)))
	ml.selected = sorted[selectedIndex.Int64()]
	//log.Info("================>", "selected", ml.selected)
	return ml.selected

}

// calculate the statedb index from key and parameter
func CalculateStateDbIndex(key string, paramIndex string) []byte {
	web3key := key + paramIndex
	hash := sha3.NewLegacyKeccak256()
	var keyIndex []byte
	hash.Write(decodeHexFromString(web3key))
	keyIndex = hash.Sum(keyIndex)
	return keyIndex
}

// decode string data to hex
func decodeHexFromString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func IncreaseHexByNum(indexKeyHash []byte, num int64) string {
	x := new(big.Int).SetBytes(indexKeyHash)
	y := big.NewInt(int64(num))
	x.Add(x, y)
	return hex.EncodeToString(x.Bytes())
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
