package main

import (
	// "encoding/base64"
	"fmt"
	log "github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"os"
	"path/filepath"
	// "strings"
)

// backupCONSUL walks an etcd tree, applying
// a reap function per leaf node visited
func backupCONSUL() bool {
	if brf.Endpoint == "" {
		return false
	}
	cfg := consul.Config{
		Address: brf.Endpoint,
	}
	cclient, _ := consul.NewClient(&cfg)
	ckv = cclient.KV()
	// use the Consul API to visit each node and store
	// the values in the local filesystem:
	visitCONSUL("/", reapsimple)
	// if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
	// 	// create an archive file of the node's values:
	// 	res := arch()
	// 	// transfer to remote, if applicable:
	// 	toremote(res)
	// }
	return true
}

// visitCONSUL visits a path in the Consul K/V tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitCONSUL(path string, fn reap) {
	log.WithFields(log.Fields{"func": "visitCONSUL"}).Debug(fmt.Sprintf("On node %s", path))
	qopts := consul.QueryOptions{
		RequireConsistent: true,
	}
	if children, _, err := ckv.Keys(path, ",", &qopts); err != nil {
		log.WithFields(log.Fields{"func": "visitCONSUL"}).Error(fmt.Sprintf("%s", err))
	} else {
		if len(children) > 1 { // there are children
			log.WithFields(log.Fields{"func": "visitCONSUL"}).Debug(fmt.Sprintf("%s has %d children", path, len(children)))
			for _, node := range children {
				log.WithFields(log.Fields{"func": "visitCONSUL"}).Debug(fmt.Sprintf("Next visiting child %s", node))
				visitCONSUL(node, fn)
			}
		} else { // we're on a leaf node
			if node, _, err := ckv.Get(path, &qopts); err != nil {
				log.WithFields(log.Fields{"func": "visitCONSUL"}).Error(fmt.Sprintf("%s", err))
			} else {
				fn("/"+node.Key, string(node.Value))
			}
		}
	}
}

func restoreCONSUL() bool {
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// transfer from remote, if applicable:
		a := fromremote()
		// unarchive:
		s := unarch(a)
		defer func() {
			_ = os.RemoveAll(s)
		}()
		cfg := consul.Config{
			Address: "http://" + brf.Endpoint,
		}
		cclient, _ := consul.NewClient(&cfg)
		ckv = cclient.KV()
		// walk the snapshot directory and use the etcd API to
		// restore keys from the local filesystem - note that
		// only non-existing keys will be created:
		if err := filepath.Walk(s, visitETCDReverse); err != nil {
			log.WithFields(log.Fields{"func": "restoreETCD"}).Error(fmt.Sprintf("%s", err))
			return false
		}
	} else { // can't restore from TTY
		return false
	}
	return true
}

func visitCONSULReverse(path string, f os.FileInfo, err error) error {
	// if f.Name() == BURRYMETA_FILE || f.Name() == snapshotid {
	// 	return nil
	// } else {
	// 	cwd, _ := os.Getwd()
	// 	base, _ := filepath.Abs(filepath.Join(cwd, snapshotid))
	// 	key, _ := filepath.Rel(base, path)
	// 	// append the root "/" to make it a key and unescape ":"
	// 	key = "/" + strings.Replace(key, "BURRY_ESC_COLON", ":", -1)
	// 	if f.IsDir() {
	// 		cfile, _ := filepath.Abs(filepath.Join(path, CONTENT_FILE))
	// 		if _, eerr := os.Stat(cfile); eerr == nil { // there is a content file at this path
	// 			log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as leaf key", key))
	// 			if c, cerr := readc(cfile); cerr != nil {
	// 				log.WithFields(log.Fields{"func": "visitETCDReverse"}).Error(fmt.Sprintf("%s", cerr))
	// 				return cerr
	// 			} else {
	// 				if _, kerr := kapi.Set(context.Background(), key, string(c), &etcd.SetOptions{Dir: false, PrevExist: etcd.PrevNoExist}); kerr == nil {
	// 					log.WithFields(log.Fields{"func": "visitETCDReverse"}).Info(fmt.Sprintf("Restored %s", key))
	// 					log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Value: %s", c))
	// 					numrestored = numrestored + 1
	// 				}
	// 			}
	// 		} else {
	// 			log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as a non-leaf key", key))
	// 			if _, kerr := kapi.Set(context.Background(), key, "", &etcd.SetOptions{Dir: true, PrevExist: etcd.PrevNoExist}); kerr == nil {
	// 				log.WithFields(log.Fields{"func": "visitETCDReverse"}).Info(fmt.Sprintf("Restored %s", key))
	// 				numrestored = numrestored + 1
	// 			}
	// 		}
	// 	}
	// 	log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Visited %s", key))
	// }
	return nil
}