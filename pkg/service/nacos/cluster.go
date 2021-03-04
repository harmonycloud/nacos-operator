package nacosClient

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type INacosClient interface {
}

type NacosClient struct {
	logger     log.Logger
	httpClient http.Client
}

type NacosClusterNodes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		State      string `json:"state"`
		ExtendInfo struct {
			AdWeight        string `json:"adWeight"`
			LastRefreshTime int64  `json:"lastRefreshTime"`
			RaftMetaData    struct {
				MetaDataMap struct {
					NamingPersistentService struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service"`
				} `json:"metaDataMap"`
			} `json:"raftMetaData"`
			RaftPort string `json:"raftPort"`
			Site     string `json:"site"`
			Version  string `json:"version"`
			// bug 有些是int，有些是string
			//Weight   string `json:"weight"`
		} `json:"extendInfo"`
		Address       string `json:"address"`
		FailAccessCnt int    `json:"failAccessCnt"`
	} `json:"data"`
}

func (c *NacosClient) GetClusterNodes(ip string) (string, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("http://%s:8848/nacos/v1/core/cluster/nodes", ip))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//func (c *CheckClient) getClusterNodesStaus(ip string) (bool, error) {
//	str, err := c.getClusterNodes(ip)
//	if err != nil {
//		return false, err
//	}
//
//}
