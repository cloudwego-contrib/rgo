/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package db

import (
	"context"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type ServerManage struct {
	gorm.Model
	ServiceName   string
	IdlContent    string
	ServerVersion string
}

func (sm *ServerManage) TableName() string {
	return "server_management"
}

func init() {
	dsn := os.Getenv("RGO_MYSQL_DSN")
	if dsn == "" {
		dsn = "gorm:gorm@tcp(localhost:3306)/gorm?charset=utf8&parseTime=True&loc=Local"
	}
	sqlDb, err := gorm.Open(mysql.Open(dsn),
		&gorm.Config{
			PrepareStmt: true,
		},
	)
	if err != nil {
		panic(err)
	}
	db = sqlDb
}

func CreateSM(ctx context.Context, sm *ServerManage) error {
	if err := db.WithContext(ctx).Create(sm).Error; err != nil {
		return err
	}
	return nil
}

func GetSMByServiceNameVersion(ctx context.Context, name, version string, sm *ServerManage) error {
	if err := db.WithContext(ctx).Where("service_name = ? and server_version = ?", name, version).
		First(sm).Error; err != nil {
		return err
	}
	return nil
}

func GetLastSMByServiceName(ctx context.Context, name string, sm *ServerManage) error {
	if err := db.WithContext(ctx).Where("service_name = ?", name).Last(sm).Error; err != nil {
		return err
	}
	return nil
}

func MGetSMByServiceName(ctx context.Context, name string, sm *[]*ServerManage) error {
	if err := db.WithContext(ctx).Where("service_name = ?", name).Find(sm).Error; err != nil {
		return err
	}
	return nil
}
