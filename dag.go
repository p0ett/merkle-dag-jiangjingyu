package merkledag

import (
	"hash"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// Add 将Node中的数据保存在KVStore中，并返回计算出的Merkle Root
func Add(kvstore KVStore, node Node) (string, error) {
	// 1. 将Node中的数据保存在KVStore中
	data, err := serialize(node)
	if err != nil {
		return "", err
	}

	key := generateKey(node)
	err = kvstore.Put(key, data)
	if err != nil {
		return "", err
	}

	// 2. 计算Merkle Root
	merkleRoot, err := calculateMerkleRoot(kvstore, node)
	if err != nil {
		return "", err
	}

	return merkleRoot, nil
}

// serialize 将Node中的数据序列化为字节数组
func serialize(node Node) ([]byte, error) {
	switch n := node.(type) {
	case File:
		return n.Bytes(), nil
	case Dir:
		it := n.It()
		var serializedData []byte
		for it.Next() {
			childNode := it.Node()
			childData, err := serialize(childNode)
			if err != nil {
				return nil, err
			}
			serializedData = append(serializedData, childData...)
		}
		return serializedData, nil
	default:
		return nil, errors.New("unsupported node type")
	}
}

// generateKey 根据Node生成唯一的存储键值
func generateKey(node Node) string {
	switch node.Type() {
	case FILE:
		fileNode := node.(File)
		return "file_" + hex.EncodeToString(fileNode.Bytes())
	case DIR:
		dirNode := node.(Dir)
		return "dir_" + hex.EncodeToString([]byte(dirNode.Size())) // 使用文件夹的大小作为键值
	default:
		return "unknown"
	}
}

// calculateMerkleRoot 计算Merkle Root
func calculateMerkleRoot(hashes []string) (string, error) {
    if len(hashes) == 0 {
        return "", errors.New("no hashes provided")
    }
    if len(hashes) == 1 {
        return hashes[0], nil
    }

    // 逐层计算Merkle Root
    for len(hashes) > 1 {
        // 如果哈希列表长度为奇数，则将最后一个哈希复制一份并添加到列表中
        if len(hashes)%2 != 0 {
            hashes = append(hashes, hashes[len(hashes)-1])
        }

        var newHashes []string
        // 两两组合计算
        for i := 0; i < len(hashes); i += 2 {
            combinedHash := sha256.New()
            combinedHash.Write([]byte(hashes[i] + hashes[i+1]))
            newHash := hex.EncodeToString(combinedHash.Sum(nil))
            newHashes = append(newHashes, newHash)
        }

        // 更新哈希列表
        hashes = newHashes
    }
	
    // 最终列表中的唯一元素即为Merkle Root
    return hashes[0], nil
}