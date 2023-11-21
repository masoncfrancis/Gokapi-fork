package database

import (
	"github.com/forceu/gokapi/internal/models"
	"github.com/forceu/gokapi/internal/test"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Setenv("GOKAPI_CONFIG_DIR", "test")
	os.Setenv("GOKAPI_DATA_DIR", "test")
	os.Mkdir("test", 0777)
	exitVal := m.Run()
	os.RemoveAll("test")
	os.Exit(exitVal)
}

func TestInit(t *testing.T) {
	Init("./test", "gokapi.sqlite")
	test.IsEqualBool(t, sqliteDb != nil, true)
	// Test that second init doesn't raise an error
	Init("./test", "gokapi.sqlite")
}

func TestClose(t *testing.T) {
	test.IsEqualBool(t, sqliteDb != nil, true)
	Close()
	test.IsEqualBool(t, sqliteDb == nil, true)
	Init("./test", "gokapi.sqlite")
}

func TestMetaData(t *testing.T) {
	files := GetAllMetadata()
	test.IsEqualInt(t, len(files), 0)

	SaveMetaData(models.File{Id: "testfile", Name: "test.txt", ExpireAt: time.Now().Add(time.Hour).Unix()})
	files = GetAllMetadata()
	test.IsEqualInt(t, len(files), 1)
	test.IsEqualString(t, files["testfile"].Name, "test.txt")

	file, ok := GetMetaDataById("testfile")
	test.IsEqualBool(t, ok, true)
	test.IsEqualString(t, file.Id, "testfile")
	_, ok = GetMetaDataById("invalid")
	test.IsEqualBool(t, ok, false)

	test.IsEqualInt(t, len(GetAllMetadata()), 1)
	DeleteMetaData("invalid")
	test.IsEqualInt(t, len(GetAllMetadata()), 1)
	DeleteMetaData("testfile")
	test.IsEqualInt(t, len(GetAllMetadata()), 0)
}

func TestHotlink(t *testing.T) {
	SaveHotlink(models.File{Id: "testhfile", Name: "testh.txt", HotlinkId: "testlink", ExpireAt: time.Now().Add(time.Hour).Unix()})

	hotlink, ok := GetHotlink("testlink")
	test.IsEqualBool(t, ok, true)
	test.IsEqualString(t, hotlink, "testhfile")
	_, ok = GetHotlink("invalid")
	test.IsEqualBool(t, ok, false)

	DeleteHotlink("invalid")
	_, ok = GetHotlink("testlink")
	test.IsEqualBool(t, ok, true)
	DeleteHotlink("testlink")
	_, ok = GetHotlink("testlink")
	test.IsEqualBool(t, ok, false)

	SaveHotlink(models.File{Id: "testhfile", Name: "testh.txt", HotlinkId: "testlink", ExpireAt: 0, UnlimitedTime: true})
	hotlink, ok = GetHotlink("testlink")
	test.IsEqualBool(t, ok, true)
	test.IsEqualString(t, hotlink, "testhfile")
}

func TestApiKey(t *testing.T) {
	SaveApiKey(models.ApiKey{
		Id:             "newkey",
		FriendlyName:   "New Key",
		LastUsed:       100,
		LastUsedString: "LastUsed",
	})
	SaveApiKey(models.ApiKey{
		Id:             "newkey2",
		FriendlyName:   "New Key2",
		LastUsed:       200,
		LastUsedString: "LastUsed2",
	})

	keys := GetAllApiKeys()
	test.IsEqualInt(t, len(keys), 2)
	test.IsEqualString(t, keys["newkey"].FriendlyName, "New Key")
	test.IsEqualString(t, keys["newkey"].Id, "newkey")
	test.IsEqualString(t, keys["newkey"].LastUsedString, "LastUsed")
	test.IsEqualBool(t, keys["newkey"].LastUsed == 100, true)

	test.IsEqualInt(t, len(GetAllApiKeys()), 2)
	DeleteApiKey("newkey2")
	test.IsEqualInt(t, len(GetAllApiKeys()), 1)

	key, ok := GetApiKey("newkey")
	test.IsEqualBool(t, ok, true)
	test.IsEqualString(t, key.FriendlyName, "New Key")
	_, ok = GetApiKey("newkey2")
	test.IsEqualBool(t, ok, false)

	SaveApiKey(models.ApiKey{
		Id:             "newkey",
		FriendlyName:   "Old Key",
		LastUsed:       100,
		LastUsedString: "LastUsed",
	})
	key, ok = GetApiKey("newkey")
	test.IsEqualBool(t, ok, true)
	test.IsEqualString(t, key.FriendlyName, "Old Key")
}

func TestSession(t *testing.T) {
	renewAt := time.Now().Add(1 * time.Hour).Unix()
	SaveSession("newsession", models.Session{
		RenewAt:    renewAt,
		ValidUntil: time.Now().Add(2 * time.Hour).Unix(),
	})

	session, ok := GetSession("newsession")
	test.IsEqualBool(t, ok, true)
	test.IsEqualBool(t, session.RenewAt == renewAt, true)

	DeleteSession("newsession")
	_, ok = GetSession("newsession")
	test.IsEqualBool(t, ok, false)

	SaveSession("newsession", models.Session{
		RenewAt:    renewAt,
		ValidUntil: time.Now().Add(2 * time.Hour).Unix(),
	})

	SaveSession("anothersession", models.Session{
		RenewAt:    renewAt,
		ValidUntil: time.Now().Add(2 * time.Hour).Unix(),
	})
	_, ok = GetSession("newsession")
	test.IsEqualBool(t, ok, true)
	_, ok = GetSession("anothersession")
	test.IsEqualBool(t, ok, true)

	DeleteAllSessions()
	_, ok = GetSession("newsession")
	test.IsEqualBool(t, ok, false)
	_, ok = GetSession("anothersession")
	test.IsEqualBool(t, ok, false)
}

func TestUploadDefaults(t *testing.T) {
	defaults := GetUploadDefaults()
	test.IsEqualInt(t, defaults.Downloads, 1)
	test.IsEqualInt(t, defaults.TimeExpiry, 14)
	test.IsEqualString(t, defaults.Password, "")
	test.IsEqualBool(t, defaults.UnlimitedDownload, false)
	test.IsEqualBool(t, defaults.UnlimitedTime, false)

	SaveUploadDefaults(models.LastUploadValues{
		Downloads:         20,
		TimeExpiry:        30,
		Password:          "abcd",
		UnlimitedDownload: true,
		UnlimitedTime:     true,
	})
	defaults = GetUploadDefaults()
	test.IsEqualInt(t, defaults.Downloads, 20)
	test.IsEqualInt(t, defaults.TimeExpiry, 30)
	test.IsEqualString(t, defaults.Password, "abcd")
	test.IsEqualBool(t, defaults.UnlimitedDownload, true)
	test.IsEqualBool(t, defaults.UnlimitedTime, true)
}

func TestGarbageCollectionUploads(t *testing.T) {
	orgiginalFunc := currentTime
	currentTime = func() time.Time {
		return time.Now().Add(-25 * time.Hour)
	}
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctodelete1",
		CurrentStatus: 0,
		LastUpdate:    time.Now().Add(-24 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctodelete2",
		CurrentStatus: 1,
		LastUpdate:    time.Now().Add(-24 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctodelete3",
		CurrentStatus: 0,
		LastUpdate:    0,
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctodelete4",
		CurrentStatus: 0,
		LastUpdate:    time.Now().Add(-20 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctodelete5",
		CurrentStatus: 1,
		LastUpdate:    time.Now().Add(40 * time.Hour).Unix(),
	})
	currentTime = orgiginalFunc

	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctokeep1",
		CurrentStatus: 0,
		LastUpdate:    time.Now().Add(-24 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctokeep2",
		CurrentStatus: 1,
		LastUpdate:    time.Now().Add(-24 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctokeep3",
		CurrentStatus: 0,
		LastUpdate:    0,
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctokeep4",
		CurrentStatus: 0,
		LastUpdate:    time.Now().Add(-20 * time.Hour).Unix(),
	})
	SaveUploadStatus(models.UploadStatus{
		ChunkId:       "ctokeep5",
		CurrentStatus: 1,
		LastUpdate:    time.Now().Add(40 * time.Hour).Unix(),
	})
	for _, item := range []string{"ctodelete1", "ctodelete2", "ctodelete3", "ctodelete4", "ctokeep1", "ctokeep2", "ctokeep3", "ctokeep4"} {
		_, result := GetUploadStatus(item)
		test.IsEqualBool(t, result, true)
	}
	RunGarbageCollection()
	for _, item := range []string{"ctodelete1", "ctodelete2", "ctodelete3", "ctodelete4"} {
		_, result := GetUploadStatus(item)
		test.IsEqualBool(t, result, false)
	}
	for _, item := range []string{"ctokeep1", "ctokeep2", "ctokeep3", "ctokeep4"} {
		_, result := GetUploadStatus(item)
		test.IsEqualBool(t, result, true)
	}
}

func TestGarbageCollectionSessions(t *testing.T) {
	SaveSession("todelete1", models.Session{
		RenewAt:    time.Now().Add(-10 * time.Second).Unix(),
		ValidUntil: time.Now().Add(-10 * time.Second).Unix(),
	})
	SaveSession("todelete2", models.Session{
		RenewAt:    time.Now().Add(10 * time.Second).Unix(),
		ValidUntil: time.Now().Add(-10 * time.Second).Unix(),
	})
	SaveSession("tokeep1", models.Session{
		RenewAt:    time.Now().Add(-10 * time.Second).Unix(),
		ValidUntil: time.Now().Add(10 * time.Second).Unix(),
	})
	SaveSession("tokeep2", models.Session{
		RenewAt:    time.Now().Add(10 * time.Second).Unix(),
		ValidUntil: time.Now().Add(10 * time.Second).Unix(),
	})
	for _, item := range []string{"todelete1", "todelete2", "tokeep1", "tokeep2"} {
		_, result := GetSession(item)
		test.IsEqualBool(t, result, true)
	}
	RunGarbageCollection()
	for _, item := range []string{"todelete1", "todelete2"} {
		_, result := GetSession(item)
		test.IsEqualBool(t, result, false)
	}
	for _, item := range []string{"tokeep1", "tokeep2"} {
		_, result := GetSession(item)
		test.IsEqualBool(t, result, true)
	}
}
