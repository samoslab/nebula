package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveClientConfig(t *testing.T) {
	configFilePath := filepath.Join(os.TempDir(), "config-test.json")
	fmt.Printf("config %s\n", configFilePath)
	removeConfigFile(configFilePath)
	space := []ReadableSpace{
		ReadableSpace{SpaceNo: 0, Password: "abcdefg", Home: "default", Name: "default"},
		ReadableSpace{SpaceNo: 1, Password: "1243", Home: "private1", Name: "privacy space"},
	}
	clientConfig := &ClientConfig{
		NodeId:     "4e57d6415305f079ebc8d644d85048252ac6e86a",
		PublicKey:  "3082010a0282010100c798aa4ae73808ed6b96a5afc84c4626aa84cb6e5a64b772b62faa92c2a0fbbeebf9d1b5711b212a8484e1ac053f816bfa213cd619f12932d9f52d9535b997f9834604d65b6eb1c40b28e8c69f30a88a3afe6eb33f78ca0829467ad6701f0e48cda70c354bf1acc5db88de77eae37172080b6e2188fa59d450160190e215202a063bff5e751ec10e256ed5a17ae8b7946f427d6763fd7b0422b39e7aabf9c7fba52356a5af589bf1b99e33f8417227079380f7fb38ed76a7ccdd2ce8abf5341f99efb011bc8e86b4584be7b6fd0c297c50de14cff9222a0049d6875866018541765b9350dbd6728fb1ad150f8fcf501e5511240c722e01f6b39f617345fac3a70203010001",
		PrivateKey: "308204a30201000282010100c798aa4ae73808ed6b96a5afc84c4626aa84cb6e5a64b772b62faa92c2a0fbbeebf9d1b5711b212a8484e1ac053f816bfa213cd619f12932d9f52d9535b997f9834604d65b6eb1c40b28e8c69f30a88a3afe6eb33f78ca0829467ad6701f0e48cda70c354bf1acc5db88de77eae37172080b6e2188fa59d450160190e215202a063bff5e751ec10e256ed5a17ae8b7946f427d6763fd7b0422b39e7aabf9c7fba52356a5af589bf1b99e33f8417227079380f7fb38ed76a7ccdd2ce8abf5341f99efb011bc8e86b4584be7b6fd0c297c50de14cff9222a0049d6875866018541765b9350dbd6728fb1ad150f8fcf501e5511240c722e01f6b39f617345fac3a70203010001028201005a0d06c7c48a037d1a8d5d3371aaf7fb79f36fd4d9f396d0aa61d7135fbc41d8619ed47e8809356d795c7a74a1e984fab25f4c934c2101f56f60aeb0230d1903f9b61d7898c7d86c1a9cda68c269b1a0abfd1ef4c02e3ef86fa6dcc4e2d12020b8a82559e669a309a1829edc7b9d2211d08f57a0c9ac29db2046cd6e1092776fb105768c23475ecfab50c684bad172bfa820d5b91c2f89450df47d748aa6f75e2f39437967b6704130723fddc3f3d9c7d02f05c71a72f87bb0855b0a5b05ed12f1cfa4941abe4c2f4497b77f7c89cb2e04e9e6f955d4d05356bf77c53edd46c49b5fb9da9615b4af25e95aaba0c88eb6ce1b172f9ed3f27c2e5801a40fdaf40102818100c93354d3a683375be2443be656dadf1286c335ac2ce8037017bd161eb8dd67b28279f3232126252017b2c56a7da69fd50fe54f7b840d8d22ea6cacec8a00344da5cb796164d7c84a8533c3aa5d11143087887851c292637d71398c28c1744651c8569c330cad53ff2039463a06eacd10e6a3b8e14971ec4813744edcf9d1e1a302818100fdf57bd828c24ec2cc4e2fb5969038c662acbc6ca7377f6f453362e0ddd0ffd354542ba8f82c2950cb963bfda999668acf0bf3f19320067ba9e93245ea67a167d9cc78519862bc5b45bc3125a9dfdd66cc73b51f7ccfb8dd61decb7dda8e4e28752f72bb97e2a763746d274f7a968661f607b172b9e19d4cd14f3a846a171e2d028181009eff2f93aa4c9b5cc32c04e1fbd52edf671bb0b8852c3c3b42a72c69ced138773ae0c0210cbb262f7c4acbf361a4613e7037585e55769807f59537fa1cfc18591c21f5a3df9b1e2eb5a0b88952ce8253ef670b1e2152a9c8a1c7465996b71a32dacc86d758b7485f9ec96413cb0f964a3ad93aeaddad677975d63dd4269f935d02818019e8976fe008a2bc60d7812a8767c3430a02115f0c582f1a0cf7471925c812b15ea30fa937585a06b21e6b945f5f1505084671e6ad59d10f80b017bee64118485e01ec2c76dd6bd5ebf15d1a38906c27f6a7bf4cad110c0d19d4fef1a2006e9cd607b72ec83a0955ae250ca3a1200629ac4df09e81b430b60b8c87adc69d01290281802a916bcbf55288ce37293d118a7aa111af423ab8c41d7928081426bde8e002f67b4bc0ed1bd68dadb82e061f31b77a03c4e7a91acd23af1ac8ace6cbef690f956142ea8fac7aacc5f48e0f64b5ca2a7b2cd8fb243788d99819247b1c55ee7939d7956c7cc5743e9e1f3dedb8df8159c4536140f6e1897015ca19c66dcbee3002",
		Email:      "test-email",
		Root:       "D:\\abc",
		Space:      space,
	}
	err := SaveClientConfig(configFilePath, clientConfig)
	require.NoError(t, err)

	size := getConfigFileSize(configFilePath)
	fmt.Printf("size %d\n", size)
	require.True(t, size > 100)

	cc, err := LoadConfig(configFilePath)
	require.NotNil(t, cc)
	require.Equal(t, cc.NodeId, clientConfig.NodeId)
	require.Equal(t, cc.PublicKey, clientConfig.PublicKey)
	fmt.Printf("space %+v\n", cc.Space)
	removeConfigFile(configFilePath)
}

func removeConfigFile(filename string) {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return
	}
	err = os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

func getConfigFileSize(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}
