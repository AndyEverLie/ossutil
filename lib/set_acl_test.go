package lib 

import (
    "fmt"
    "os"
    "strconv"

    . "gopkg.in/check.v1"
)

func (s *OssutilCommandSuite) rawSetBucketACL(bucket, acl string, force bool) (bool, error) {
    args := []string{CloudURLToString(bucket, ""), acl}
    showElapse, err := s.rawSetACLWithArgs(args, false, true, force)
    return showElapse, err
}

func (s *OssutilCommandSuite) rawSetObjectACL(bucket, object, acl string, recursive, force bool) (bool, error) {
    args := []string{CloudURLToString(bucket, object), acl}
    showElapse, err := s.rawSetACLWithArgs(args, recursive, false, force)
    return showElapse, err
}

func (s *OssutilCommandSuite) rawSetACLWithArgs(args []string, recursive, bucket, force bool) (bool, error) {
    command := "set-acl"
    str := ""
    routines := strconv.Itoa(Routines)
    options := OptionMapType{
        "endpoint": &str,
        "accessKeyID": &str,
        "accessKeySecret": &str,
        "stsToken": &str,
        "configFile": &configFile,
        "routines": &routines,
        "recursive": &recursive,
        "bucket": &bucket,
        "force": &force,
    }
    showElapse, err := cm.RunCommand(command, args, options)
    return showElapse, err
}

func (s *OssutilCommandSuite) TestSetBucketACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    // get acl
    bucketStat := s.getStat(bucket, "", c)
    c.Assert(bucketStat[StatACL], Equals, "private")

    // set acl
    for _, acl := range []string{"private", "public-read", "public-read-write"} {
        s.setBucketACL(bucket, acl, c)
        bucketStat := s.getStat(bucket, "", c)
        c.Assert(bucketStat[StatACL], Equals, acl)
    }

    result := []string{"private", "public-read", "public-read-write"}
    for i, acl := range []string{"private", "public-read", "public-read-write"} {
        s.setBucketACL(bucket, acl, c)
        bucketStat := s.getStat(bucket, "", c)
        c.Assert(bucketStat[StatACL], Equals, result[i])
    }
}

func (s *OssutilCommandSuite) TestSetBucketErrorACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    for _, acl := range []string{"default", "def", "erracl", "私有"} {
        showElapse, err := s.rawSetBucketACL(bucket, acl, false)
        c.Assert(err, NotNil)
        c.Assert(showElapse, Equals, false)

        showElapse, err = s.rawSetBucketACL(bucket, acl, true)
        c.Assert(err, NotNil)
        c.Assert(showElapse, Equals, false)

        bucketStat := s.getStat(bucket, "", c)
        c.Assert(bucketStat[StatACL], Equals, "private")
    }
}

func (s *OssutilCommandSuite) TestSetNotExistBucketACL(c *C) {
    bucket := bucketNamePrefix + "notexist"

    showElapse, err := s.rawGetStat(bucket, "")
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    // set acl and create bucket
    showElapse, err = s.rawSetBucketACL(bucket, "public-read", true)
    c.Assert(err, IsNil)
    c.Assert(showElapse, Equals, true)

    bucketStat := s.getStat(bucket, "", c)
    c.Assert(bucketStat[StatACL], Equals, "public-read")

    // invalid bucket name
    bucket = "a"
    showElapse, err = s.rawSetBucketACL(bucket, "public-read", true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    showElapse, err = s.rawGetStat(bucket, "")
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)
}

func (s *OssutilCommandSuite) TestSetBucketEmptyACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    object := "test"
    s.putObject(bucket, object, uploadFileName, c)

    showElapse, err := s.rawSetObjectACL(bucket, object, "", false, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)
}

func (s *OssutilCommandSuite) TestSetObjectACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    object := "文件"

    // set acl to not exist object
    showElapse, err := s.rawSetObjectACL(bucket, object, "default", false, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    s.putObject(bucket, object, uploadFileName, c)

    //get object acl
    objectStat := s.getStat(bucket, object, c)
    c.Assert(objectStat[StatACL], Equals, "default")

    // set acl
    for _, acl := range []string{"default", "private", "public-read", "public-read-write"} {
        s.setObjectACL(bucket, object, acl, false, true, c)
        objectStat = s.getStat(bucket, object, c)
        c.Assert(objectStat[StatACL], Equals, acl)
    }

    result := []string{"default", "private", "public-read", "public-read-write"}
    for i, acl := range []string{"default", "private", "public-read", "public-read-write"} {
        s.setObjectACL(bucket, object, acl, false, false, c)
        objectStat = s.getStat(bucket, object, c)
        c.Assert(objectStat[StatACL], Equals, result[i])
    }


    s.setObjectACL(bucket, object, "private", false, true, c)

    // set error acl
    for _, acl := range []string{"public_read", "erracl", "私有"} {
        showElapse, err = s.rawSetObjectACL(bucket, object, acl, false, false)
        c.Assert(showElapse, Equals, false)
        c.Assert(err, NotNil)

        objectStat = s.getStat(bucket, object, c)
        c.Assert(objectStat[StatACL], Equals, "private")
    }
}

func (s *OssutilCommandSuite) TestBatchSetObjectACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    // put objects
    num := 10
    objectNames := []string{}
    for i := 0; i < num; i++ {
        object := fmt.Sprintf("设置权限:%d", i)
        s.putObject(bucket, object, uploadFileName, c)
        objectNames = append(objectNames, object)
    }

    // without --force option
    s.setObjectACL(bucket, "", "public-read-write", true, false, c)

    for _, object := range objectNames {
        objectStat := s.getStat(bucket, object, c)
        c.Assert(objectStat[StatACL], Equals, "default")
    }

    for _, acl := range []string{"default", "private", "public-read", "public-read-write"} {
        s.setObjectACL(bucket, "设置权限", acl, true, true, c)

        for _, object := range objectNames {
            objectStat := s.getStat(bucket, object, c)
            c.Assert(objectStat[StatACL], Equals, acl)
        }
    }

    showElapse, err := s.rawSetObjectACL(bucket, "设置权限", "erracl", true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)
}

func (s *OssutilCommandSuite) TestErrSetACL(c *C) {
    acl := "private"
    args := []string{"os://", acl}
    showElapse, err := s.rawSetACLWithArgs(args, false, false, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    args = []string{"oss://", acl} 
    showElapse, err = s.rawSetACLWithArgs(args, false, false, true)

    // error set bucket acl
    bucket := bucketNamePrefix + "set-acl"
    object := "testobject"
    args = []string{CloudURLToString(bucket, object), acl}
    showElapse, err = s.rawSetACLWithArgs(args, false, true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    args = []string{CloudURLToString(bucket, ""), acl}
    showElapse, err = s.rawSetACLWithArgs(args, true, true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    // miss acl
    args = []string{CloudURLToString(bucket, "")}
    showElapse, err = s.rawSetACLWithArgs(args, false, true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    args = []string{CloudURLToString(bucket, object)}
    showElapse, err = s.rawSetACLWithArgs(args, false, true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    args = []string{CloudURLToString(bucket, object)}
    showElapse, err = s.rawSetACLWithArgs(args, true, false, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    // miss object
    args = []string{CloudURLToString(bucket, ""), acl}
    showElapse, err = s.rawSetACLWithArgs(args, false, false, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    // bad prefix
    showElapse, err = s.rawSetObjectACL(bucket, "/object", acl, true, true)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)
}

func (s *OssutilCommandSuite) TestErrBatchSetACL(c *C) {
    bucket := bucketNamePrefix + "set-acl"
    s.putBucket(bucket, c)

    // put objects
    num := 10
    objectNames := []string{}
    for i := 0; i < num; i++ {
        object := fmt.Sprintf("设置权限:%d", i)
        s.putObject(bucket, object, uploadFileName, c)
        objectNames = append(objectNames, object)
    }

    command := "set-acl"
    str := ""
    str1 := "abc"
    args := []string{CloudURLToString(bucket, ""), "public-read-write"}
    routines := strconv.Itoa(Routines)
    ok := true
    options := OptionMapType{
        "endpoint": &str1,
        "accessKeyID": &str1,
        "accessKeySecret": &str1,
        "stsToken": &str,
        "routines": &routines,
        "recursive": &ok,
        "force": &ok,
    }
    showElapse, err := cm.RunCommand(command, args, options)
    c.Assert(err, NotNil)
    c.Assert(showElapse, Equals, false)

    for _, object := range objectNames {
        objectStat := s.getStat(bucket, object, c)
        c.Assert(objectStat[StatACL], Equals, "default")
    }
}

func (s *OssutilCommandSuite) TestSetACLIDKey(c *C) {
    bucket := bucketNamePrefix + "set-aclidkey"
    s.putBucket(bucket, c)

    cfile := "ossutil_test.config_boto"
    data := fmt.Sprintf("[Credentials]\nendpoint=%s\naccessKeyID=%s\naccessKeySecret=%s\n[Bucket-Endpoint]\n%s=%s[Bucket-Cname]\n%s=%s", "abc", "def", "ghi", bucket, "abc", bucket, "abc") 
    s.createFile(cfile, data, c)

    command := "set-acl"
    str := ""
    args := []string{CloudURLToString(bucket, ""), "public-read"}
    routines := strconv.Itoa(Routines)
    ok := true
    options := OptionMapType{
        "endpoint": &str,
        "accessKeyID": &str,
        "accessKeySecret": &str,
        "stsToken": &str,
        "configFile": &cfile,
        "routines": &routines,
        "bucket": &ok,
        "force": &ok,
    }
    showElapse, err := cm.RunCommand(command, args, options)
    c.Assert(err, NotNil)

    options = OptionMapType{
        "endpoint": &endpoint,
        "accessKeyID": &accessKeyID,
        "accessKeySecret": &accessKeySecret,
        "stsToken": &str,
        "configFile": &cfile,
        "routines": &routines,
        "bucket": &ok,
        "force": &ok,
    }
    showElapse, err = cm.RunCommand(command, args, options)
    c.Assert(err, IsNil)
    c.Assert(showElapse, Equals, true)

    _ = os.Remove(cfile)
}
