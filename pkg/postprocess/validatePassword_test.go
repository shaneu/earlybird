/*
 * Copyright 2021 American Express
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package postprocess

import "testing"

type args struct {
	testPD string
}

var tests = []struct {
	name           string
	args           args
	wantConfidence int
	wantIgnore     bool
}{
	{
		name: "Skip passwords too short",
		args: args{
			testPD: "fo",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Skip variables",
		args: args{
			testPD: "$variable",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Skip functions",
		args: args{
			testPD: "func()",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Skip passwords with spaces and no quotes",
		args: args{
			testPD: "ignore me please",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Skip passwords with a dot",
		args: args{
			testPD: "password: ignore.me",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Skip passwords with two equals",
		args: args{
			testPD: "password: ignoreme==please",
		},
		wantConfidence: 3,
		wantIgnore:     true,
	},
	{
		name: "Do not skip, real finding",
		args: args{
			testPD: "VeryStrong857#",
		},
		wantConfidence: 2,
		wantIgnore:     false,
	},
	{
		name: "Verify = delimited values",
		args: args{
			testPD: "my.property=propertyEqualDelimitedPassword",
		},
		wantConfidence: 2,
		wantIgnore:     false,
	},
	{
		name: "Verify : delimited values",
		args: args{
			testPD: "my.property:propertyColonDelimitedPassword",
		},
		wantConfidence: 2,
		wantIgnore:     false,
	},
	{
		name: "Do not skip, real finding, ensure whitespace is permitted around delimited values",
		args: args{
			testPD: "my.property    =     propertySpacesAroundDelimited",
		},
		wantConfidence: 2,
		wantIgnore:     false,
	},
	{
		name: "Do not skip, real finding, ensure yml files are handled",
		args: args{
			testPD: "my.property: sampleYmlPassword",
		},
		wantConfidence: 2,
		wantIgnore:     false,
	},
	{
		name: "Do not skip, real finding, ensure json files are handled",
		args: args{
			testPD: "\"my.property\": \"sample%3YmlPassword\"",
		},
		wantConfidence: 3,
		wantIgnore:     false,
	},
}

func TestPasswordFalse(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfidence, gotIgnore := PasswordFalse(tt.args.testPD)
			if gotConfidence != tt.wantConfidence {
				t.Errorf("PasswordFalse() gotConfidence = %v, want %v", gotConfidence, tt.wantConfidence)
			}
			if gotIgnore != tt.wantIgnore {
				t.Errorf("PasswordFalse() gotIgnore = %v, want %v", gotIgnore, tt.wantIgnore)
			}
		})
	}
}

type params struct {
	matchValue string
	lineValue  string
}

var testsSkipSameKeyValuePasswords []struct {
	name       string
	params     params
	wantIgnore bool
} = []struct {
	name       string
	params     params
	wantIgnore bool
}{
	{
		name: "Skip same key/value passwords as case insensitive in properties file",
		params: params{
			matchValue: "PASSWORD_DB : password_db",
			lineValue:  "PASSWORD_DB : password_db",
		},
		wantIgnore: true,
	},
	{
		name: "Skip same key/value passwords as case insensitive in yaml file in lineValue",
		params: params{
			matchValue: "PASSWORD : db_password",
			lineValue:  "DB_PASSWORD : db_password",
		},
		wantIgnore: true,
	},
	{
		name: "Skip same key/value passwords that match alphanumerically in lineValue",
		params: params{
			matchValue: "PASSWORD : $db.password",
			lineValue:  "DB_PASSWORD : $db.password",
		},
		wantIgnore: true,
	},
	{
		name: "Skip same key/value secret in properties file",
		params: params{
			matchValue: "PASSWORD : $db.password",
			lineValue:  "DB_PASSWORD : $db.password",
		},
		wantIgnore: true,
	},
	{
		name: "Skip same key/value secret in yaml file in lineValue",
		params: params{
			matchValue: "SECRET: api.Secret",
			lineValue:  "APISECRET: api.Secret",
		},
		wantIgnore: true,
	},
	{
		name: "Skip same key/value secret in json file",
		params: params{
			matchValue: "\"SECRET\": \"SECRET\"",
			lineValue:  "\"SECRET\": \"SECRET\"",
		},
		wantIgnore: true,
	},
	{
		name: "Do not skip, real password finding",
		params: params{
			matchValue: "password_couchbase: VeryStrong857#",
			lineValue:  "password_couchbase: VeryStrong857#",
		},
		wantIgnore: false,
	},
	{
		name: "Do not skip, real password finding",
		params: params{
			matchValue: "password: VeryStrong857#",
			lineValue:  "couchbase_password: VeryStrong857#",
		},
		wantIgnore: false,
	},
	{
		name: "Do not skip, real secret finding",
		params: params{
			matchValue: "secret: VeryStrong857#",
			lineValue:  "secret: VeryStrong857#",
		},
		wantIgnore: false,
	},
	{
		name: "Skip same key/value when match value and line value are different",
		params: params{
			matchValue: "Secret=npazAppSecret",
			lineValue:  "api.appSecret=apiAppSecret",
		},
		wantIgnore: true,
	},
}

func TestSkipPasswordSameKeyValue(t *testing.T) {
	for _, tt := range testsSkipSameKeyValuePasswords {
		t.Run(tt.name, func(t *testing.T) {
			gotIgnore := SkipSameKeyValuePassword(tt.params.matchValue, tt.params.lineValue)
			if gotIgnore != tt.wantIgnore {
				t.Errorf("SkipSameKeyValuePassword() gotIgnore = %v, want %v", gotIgnore, tt.wantIgnore)
			}
		})
	}
}

var testSkipUnicodeInPasswords = []struct {
	name       string
	args       args
	wantIgnore bool
}{
	{
		name: "Skip password with unicode which is not ASCII",
		args: args{
			testPD: `"password": "\u0049\u0044\u306e\u78ba\u8a8d\u3001\u30d1\u30b9\u30ef\u30fc\u30c9\u306e\u5909\u66f4"`,
		},
		wantIgnore: true,
	},
	{
		name: "Skip passwords that has non ASCII char",
		args: args{
			testPD: `"password": "VeryStrong$$\u306e\u78ba"`,
		},
		wantIgnore: true,
	},
	{
		name: "Do not skip password with unicode that convert to valid of ASCII.",
		args: args{
			testPD: `"password": "VeryStrong$$\u0049\u0044"`,
		},
		wantIgnore: false,
	},
	{
		name: "Do not skip, real password finding",
		args: args{
			testPD: "password: VeryStrong857!@$^&*#",
		},
		wantIgnore: false,
	},
	{
		name: "Do not skip, real secret finding",
		args: args{
			testPD: "secret: VeryStrong857#",
		},
		wantIgnore: false,
	},
	{
		name: "Skip passwords that has non ASCII char and is invalid string due to unicode being in CAPS",
		args: args{
			testPD: `"password"= "Informationsb\U00e4rare"`,
		},
		wantIgnore: true,
	},
}

func TestSkipUnicodeInPasswords(t *testing.T) {
	for _, tt := range testSkipUnicodeInPasswords {
		t.Run(tt.name, func(t *testing.T) {
			gotIgnore := SkipPasswordWithUnicode(tt.args.testPD)
			if gotIgnore != tt.wantIgnore {
				t.Errorf("SkipUnicodeInPasswords() gotIgnore = %v, want %v", gotIgnore, tt.wantIgnore)
			}
		})
	}
}
