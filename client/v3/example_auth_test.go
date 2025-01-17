// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientv3_test

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/tests/v3/integration/clientv3/examples"
	"log"
)

func mockAuth() {
	fmt.Println(`etcdserver: permission denied`)
	fmt.Println(`user u permission: key "foo", range end "zoo"`)
}

func ExampleAuth() {
	examples.forUnitTestsRunInMockedContext(
		mockAuth,
		func() {
			cli, err := New(Config{
				Endpoints:   examples.exampleEndpoints(),
				DialTimeout: examples.dialTimeout,
			})
			if err != nil {
				log.Fatal(err)
			}
			defer cli.Close()

			if _, err = cli.RoleAdd(context.TODO(), "root"); err != nil {
				log.Fatal(err)
			}
			if _, err = cli.UserAdd(context.TODO(), "root", "123"); err != nil {
				log.Fatal(err)
			}
			if _, err = cli.UserGrantRole(context.TODO(), "root", "root"); err != nil {
				log.Fatal(err)
			}

			if _, err = cli.RoleAdd(context.TODO(), "r"); err != nil {
				log.Fatal(err)
			}

			if _, err = cli.RoleGrantPermission(
				context.TODO(),
				"r",   // role name
				"foo", // key
				"zoo", // range end
				PermissionType(PermReadWrite),
			); err != nil {
				log.Fatal(err)
			}
			if _, err = cli.UserAdd(context.TODO(), "u", "123"); err != nil {
				log.Fatal(err)
			}
			if _, err = cli.UserGrantRole(context.TODO(), "u", "r"); err != nil {
				log.Fatal(err)
			}
			if _, err = cli.AuthEnable(context.TODO()); err != nil {
				log.Fatal(err)
			}

			cliAuth, err := New(Config{
				Endpoints:   examples.exampleEndpoints(),
				DialTimeout: examples.dialTimeout,
				Username:    "u",
				Password:    "123",
			})
			if err != nil {
				log.Fatal(err)
			}
			defer cliAuth.Close()

			if _, err = cliAuth.Put(context.TODO(), "foo1", "bar"); err != nil {
				log.Fatal(err)
			}

			_, err = cliAuth.Txn(context.TODO()).
				If(Compare(Value("zoo1"), ">", "abc")).
				Then(OpPut("zoo1", "XYZ")).
				Else(OpPut("zoo1", "ABC")).
				Commit()
			fmt.Println(err)

			// now check the permission with the root account
			rootCli, err := New(Config{
				Endpoints:   examples.exampleEndpoints(),
				DialTimeout: examples.dialTimeout,
				Username:    "root",
				Password:    "123",
			})
			if err != nil {
				log.Fatal(err)
			}
			defer rootCli.Close()

			resp, err := rootCli.RoleGet(context.TODO(), "r")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("user u permission: key %q, range end %q\n", resp.Perm[0].Key, resp.Perm[0].RangeEnd)

			if _, err = rootCli.AuthDisable(context.TODO()); err != nil {
				log.Fatal(err)
			}
		})
	// Output: etcdserver: permission denied
	// user u permission: key "foo", range end "zoo"
}
