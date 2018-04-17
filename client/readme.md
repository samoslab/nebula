文件如果需要加密就先加密，然后计算sha1 hash，然后进行第一步，发送请求检查文件是否存在（CheckFileExist），注意如果文件size<=8k，文件数据随请求一起发送，metadata server判断如果文件已经存在，就返回done=tue，整个过程完成，不再执行第二步和第三歩

如果文件不存在，返回请求告诉使用纠删码还是多副本（根据文件大小判断），如果要使用纠删码还会返回分多少片多少个校验片，然后执行第二步（UploadFilePrepare）；如果使用多副本，还会返回应该使用多少个副本，以及用来存储这些副本的Provider，不需要执行第二步，直接跳到第三歩

如果使用纠删码，第二步发请求把按照纠删码计算获得的分片hash和片大小发送到服务器，服务器端返回用来存储分片的Provider，注意大部分Provider是对应某个hash的分片，但有少部分provider是作为备用的，可以存储多个分片其中任意一个

不管是使用纠删码还是多副本，数据存储到Provider完成后，再执行第三歩UploadFileDone把分区和分片信息传到服务器端，如果是多副本方式，分区和分片数都是1，存储的Provider(storeNodeId)是多个

```

1. register

URI:/api/register
Method:POST
Args:
   email:string
   configdir:string
   trackServer:string

2. verify email

URI:/api/verifymail
Method:POST
Args:
   code:string
   configDir:string

3. create folder

URI:/api/mkfolder
Method: POST
Args:
  path:string

4. upload file

URI:/api/upload
Method: POST
Args:
  filename:string

5. list files

URI:/api/list
Method: get
Args:
  path:string

6. download files

URI:/api/download
Method: POST
Args:
  filehash:string
  filesize:uint64
  filename:string
  folder:bool

7. remove files

URI:/api/remove
Method: POST
Args:
   path:string

 ``` 

