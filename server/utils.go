package server

import (
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

func initBackupWriter(svrConfig *toml.Tree) MsgWriter {
	var (
		err          error
		writer       MsgWriter
		backFilePath string
	)
	if backupToggle := svrConfig.GetDefault("backup-toggle", false).(bool); backupToggle {
		if backFilePath = svrConfig.GetDefault("backup-file", "").(string); len(backFilePath) == 0 {
			log.Error("backup data filePath is empty")
			return nil
		}
	}
	if len(backFilePath) != 0 {
		if writer, err = NewFileMsgWriter(backFilePath); err != nil {
			log.WithError(err).Error("create writer error")
			return nil
		}
	}
	return writer

}
