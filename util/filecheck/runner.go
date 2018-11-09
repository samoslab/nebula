package filecheck

import "sync"

type pathAndChunkSize struct {
	path      string
	chunkSize uint32
}

type metadataInfo struct {
	paramStr    string
	generator   []byte
	pubKeyBytes []byte
	random      []byte
	phi         [][]byte
	er          error
}

type GenMetadataRunner struct {
	queue  chan *pathAndChunkSize
	quit   chan bool
	result sync.Map
}

func NewRunner() *GenMetadataRunner {
	return &GenMetadataRunner{queue: make(chan *pathAndChunkSize, 4096), quit: make(chan bool, 1)}
}

func (self *GenMetadataRunner) Run() {
	for {
		select {
		case <-self.quit:
			return
		case pacs := <-self.queue:
			self.doGen(pacs)
		}
	}
}

func (self *GenMetadataRunner) Quit() {
	self.quit <- true
}

func (self *GenMetadataRunner) AddPath(path string, chunkSize uint32) {
	self.queue <- &pathAndChunkSize{path: path, chunkSize: chunkSize}
}

func (self *GenMetadataRunner) doGen(pacs *pathAndChunkSize) {
	paramStr, generator, pubKeyBytes, random, phi, er := GenMetadata(pacs.path, pacs.chunkSize)
	self.result.Store(pacs.path, &metadataInfo{paramStr: paramStr, generator: generator, pubKeyBytes: pubKeyBytes, random: random, phi: phi, er: er})
}

func (self *GenMetadataRunner) GetResult(path string) (exist bool, paramStr string, generator []byte, pubKeyBytes []byte, random []byte, phi [][]byte, er error) {
	vi, exist := self.result.Load(path)
	if exist {
		v := vi.(*metadataInfo)
		paramStr, generator, pubKeyBytes, random, phi, er = v.paramStr, v.generator, v.pubKeyBytes, v.random, v.phi, v.er
	}
	return
}

func (self *GenMetadataRunner) RemoveResult(path string) {
	self.result.Delete(path)
}
