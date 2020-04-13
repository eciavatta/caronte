package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"os"
	"time"
)

const invalidHashString = "invalid"

func Sha256Sum(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return invalidHashString, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Println("Cannot close file " + fileName)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return invalidHashString, err
	}

	return string(h.Sum(nil)), nil
}

func CustomRowID(payload uint64, timestamp time.Time) RowID {
	var key [12]byte
	binary.BigEndian.PutUint32(key[0:4], uint32(timestamp.Unix()))
	binary.BigEndian.PutUint64(key[4:12], payload)

	if oid, err := primitive.ObjectIDFromHex(hex.EncodeToString(key[:])); err == nil {
		return oid
	} else {
		log.WithError(err).Warn("failed to create object id")
		return primitive.NewObjectID()
	}
}

func NewRowID() RowID {
	return primitive.NewObjectID()
}

func EmptyRowID() RowID {
	return [12]byte{}
}

func RowIDFromHex(hex string) (RowID, error) {
	rowID, err := primitive.ObjectIDFromHex(hex)
	return rowID, err
}
