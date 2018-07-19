var refreshTransportTimer = null;
var method = {
     //判断是否注册过邮箱！
     login:function(){
        $.ajax({
            url:"/api/v1/service/status",
            method:"GET",
            success:function(res){
                if(res.status){
                   //没有注册过！跳到注册页
                   console.log(res);
                }
            }
        })
    },
    //订单使用情况
    usage:function (){
        $.ajax({
            url:"/api/v1/usage/amount",
            success:function(res){
                console.log(res);
            }
        });
    },
    //获取hashName;
    getParamsUrl:function (){
        var hashName = location.hash.split("#")[1];//路由地址
        return 	{
            path:hashName
        }
    },
    //设置密码；发送后台；进入隐私空间；
    setPrivtePsd:function (setVal,space_no){
        console.log(setVal+";"+space_no);
        let that = this;
        $.ajax({
            url:"/api/v1/space/password",
            method:'POST',
            contentType:'application/json',
            data:JSON.stringify({
                "password": setVal,
                "space_no": space_no
            }),
            success:function(res){
                console.log(res);
                if((res.code==0)&&(res.Data=="ok")){
                    $.cookie("privitePsd", 1, { expires: 365,path: '/' });
                    $("#priviteCondition").hide();
                    $("#diskDivAll").show();
                    that.priviteInit('/',space_no);
                }
            }
        });
    },

    firstInit:function(){
        clearInterval(refreshTransportTimer);
        let path = method.getParamsUrl().path;
        location.hash = path;
        console.log(path);
        let space_no = 0;
        let a = path.split(":")[0];
         //左侧选项卡
         $("#frameAsideUl li").removeClass("active");
        if(a == "myspace"){
            space_no = 0;
            // //左侧选项卡
            $("#mySpace").parent().addClass("active");
            this.init(path,space_no);
        }else if(a == "privite"){
            space_no = 1;
            //左侧选项卡
            $("#privteSpace").parent().addClass("active");
            this.init(path,space_no);
        }else if(a=="transport"){
            this.transportInit();
        }else if(a=="package"){
            this.packageInit();
        }
        
    },
    //初始化
    init:function (path,space_no){
        path = path.split(":")[1];
        if(path.length == 0){
            path = "/";
        }
        console.log(path);
        if(space_no==0){
            this.myspaceInit(path,space_no);
        }else if(space_no==1){
            this.priviteInit(path,space_no);
        }
        
    },
    //我的空间初始化
    myspaceInit:function(path,space_no){
        //外层显示
        $("#diskDiv").show();
        // 内容区显示
        $("#diskDivAll").show();

        //隐私空间隐藏
        $("#priviteCondition").hide();
        //传输列表和我的套餐隐藏
        $(".tabDiv").hide();

        list(path,space_no,1000,1,'modtime',true);  
    },
    //隐私空间初始化
    priviteInit:function(path,space_no){
        //外层显示
        $("#diskDiv").show();
        //传输列表和我的套餐隐藏
        $(".tabDiv").hide();

        if(sessionStorage['hadImport']){
            $("#priviteCondition").hide();
            $("#diskDivAll").show();
            list(path,space_no,1000,1,'modtime',true);   
        }else{
            // 内容区隐藏
            $("#diskDivAll").hide();
            //隐私空间
            $("#priviteCondition").show();
        }

        if(!$.cookie("privitePsd")||$.cookie("privitePsd")=='0'){
            //设置密码显示
            $(".priviteConditionContentBox").eq(0).show();
            $(".priviteConditionContentBox").eq(1).hide();
        }else{
            $(".priviteConditionContentBox").eq(1).show();
            $(".priviteConditionContentBox").eq(0).hide();
        }
    },
    //传输列表初始化
    transportInit:function(){
        $("#transport").parent().addClass("active");
        $("#diskDiv").hide();
        $("#PackageDiv").hide();
        $("#transportDiv").show();
        //调用传输列表初始化的方法；
        refreshTransportTimer = setInterval(function(){
            transportMethod.transportInit();
        },2000);
         
    },
    //我的套餐初始化
    packageInit:function(){
        $("#package").parent().addClass("active");
        $("#diskDiv").hide();
        $("#transportDiv").hide();
        $("#PackageDiv").show();
        //调用套餐初始的方法们；
        packageMethod.packageInit();
    }

};


//  导出配置文件
//  $("#exportConfig").click(function(){
    // let filename = '';
    // $.ajax({
    //     url:"/api/v1/config/export",
    //     contentType:"application/json",
    //     method:'POST',
    //     data:JSON.stringify({'filename':'C:\\samos_disk/config.json'}),
    //     success:function(res){
    //         if(res.code==0){
    //             alert("The configuration file is saved and the save path is 'C:\\samos_disk/config.json'.");
    //         }else{
    //             alert("Export failed, please restart.");
    //         }
    //        // console.log(res);
    //     }
    // });
// });

//查所有文件类型
// $.ajax({
//     url:"/api/v1/service/filetype",
//     success:function(res){
//         console.log(res);
//         let arr=[];
//         $.each(res,function(index,obj){
//             arr.push(obj.extension);
//         });
//         console.log(arr);
//     }
// });

//请求列表
function list(path,space_no,pagesize,pagenum,sorttype,ascorder){
    $.ajax({
        url:"/api/v1/store/list",
        method:'POST',
        contentType: "application/json",
        data:JSON.stringify({
            "path":path,
            "space_no":space_no,
            "pagesize":pagesize,
            "pagenum":pagenum,
            "sorttype":sorttype,
            "ascorder":ascorder
        }),
        success:function(res){
            console.log(res);
            if(res.code!=0)return;
            //插入列表内容；
            append(res,path,space_no);
            //插入面包屑导航内容；
             let html = breadNav(path,space_no);
            $("#breadNav").html(html);
        }
    });
 }

 //列表中插入所有list
function append(res,path,space_no){
    let html = '';
    if((res.Data.total<1)||(res.code!=0)){
        $('.zJMtAEb').html('');
        $(".no-file-ab").show();
        return false;
    }
  

    $(".no-file-ab").hide();

    //let typeArr=["epub", "otf", "woff", "gz", "doc", "eot", "pdf", "ps", "rtf", "cab", "xls", "ppt", "pptx", "xlsx", "docx", "7z", "bz2", "Z", "deb", "elf", "crx", "lz", "exe", "nes", "rar", "rpm", "swf", "sqlite", "tar", "ar", "xz", "zip", "amr", "m4a", "mid", "mp3", "ogg", "flac", "wav", "bmp", "gif", "jpg", "png", "tif", "psd", "jxr", "webp", "cr2", "ico", "mp4", "mpg", "mov", "webm", "flv", "m4v", "mkv", "wmv", "avi"];
   
    $.each(res.Data.files,function(index,obj){
        let a = obj.extension;
            k = obj.filesize; 
            no = '';                          //文件大小
        if(k&&k<1024){
            k = obj.filesize+' B'
        }else if(k&&k>=1024&&k<1024*1024){
            k = Math.round(obj.filesize / 1024)+' KB'; 
        }else if(k&&k>=1024*1024){
            k = Math.round(obj.filesize/(1024*1024)).toFixed(2)+' M' 
        }else{
            k = '-'
        }
        if(space_no==0){
            no = '#myspace:';
        }else if(space_no==1){
            no = '#privite:'
        }
        if(path=="/"){
            path="";
        }

        html+=` <dd class="AuPKyz">
                    <div data-key="name" class="AuPKyz-li" style="width:44%;">
                        <input class="s-select-check" type="checkbox" name="fileSelect" data-name="${obj.filename}" data-id="${obj.id}" data-path="${path}${'/'+obj.filename}" data-hash="${obj.filehash}" data-size="${obj.filesize}" data-folder=${obj.folder} data-spaceNo="${space_no}">
                        <span class="file-icon my-file-${a}"></span>
                        <a class="file-name" title="${obj.filename}" href="${no}${path}${'/'+obj.filename}">${obj.filename}</a>
                    </div>
                    <div data-key="size" class="AuPKyz-li" style="width:16%;">
                        <span class="text">${k}</span>
                    </div>
                    <div data-key="size" class="AuPKyz-li" style="width:16%;">
                        <span class="text">${(obj.filetype!='unknown')?obj.filetype:'folder'}</span>
                    </div>
                    <div data-key="time" class="AuPKyz-li" style="width:23%;">
                        <span class="text">${public.Date(obj.modtime)}</span>
                    </div>
                </dd>`;
    });
    $('.zJMtAEb').html(html);

     // <!--单行选中增加类-->
    rowSelected();
    //点击文件名进入下一层
    nextLayer(space_no);

    
}


//<!--列表是否选中是否显示按钮组 全选中的情况下 全选按钮也选中  如若选中数量大于1个禁止重命名按按钮点击事件-->
function btngroupshow(){
    if($(".zJMtAEb input[name='fileSelect']:checked").size()!=0){
        $("#s-button-group").show();
    }else{
        $("#s-button-group").hide();
    }
    if($(".zJMtAEb input[name='fileSelect']:checked").size()==$(".zJMtAEb input[name='fileSelect']").length){
        $("#s-selectAll").prop('checked','checked');
    }else{
        $("#s-selectAll").prop('checked',false);
    }
    if($(".zJMtAEb input[name='fileSelect']:checked").size()!=1){
        $("#renameBtn").css('opacity',0.5);
        $("#renameBtn").click(function(event){
            event.preventDefault();
        });
    }else{
        $("#renameBtn").css('opacity',1);
    }
}

 // <!--单行选中增加类-->
 function rowSelected(){
    $(".zJMtAEb input[name='fileSelect']").each(function(){
        $(this).change(function(){
         //$(".s-select-check").change(function(){
             if($(this).is(":checked")){
                 //console.log($(this).attr("data-name"));查 选中行的文件夹名字
                 $(this).parent().parent().addClass("activeSelect");
             }else{
                 $(this).parent().parent().removeClass("activeSelect");
             }
             btngroupshow();
         });
     });
 }

 //面包屑导航 得到层级 该有的代码；
 function breadNav(path,space_no){
     let a = "";
    if(space_no==0){
        a = "#myspace:"
    }else if(space_no==1){
        a = "#privite:"
    }
    let pathArr1 = path.slice(1).split('/');
    let pathArr2 = [...pathArr1];
    let liHtml = '';
    let tmpArr = [];
    for(let i = 1;i<pathArr1.length;i++){
        let b ='';
        for(let j=0;j<i;j++){
            b+='/'+pathArr1[j]; 
        }
        tmpArr.push(b);
    }

    for(let j = 0;j<pathArr2.length;j++){
        
        liHtml+=`<li>
                    <span>></span>
                    <a href="${a}${tmpArr[j]}">${pathArr2[j]}</a>
                </li>`;
    }
    let allFileHtml =`<li>
                        <a href="${a}" title="all files"  data-locale="allFileOther">all files</a>
                    </li>`;
    allFileHtml+=liHtml;
    return allFileHtml;
 }

 //点击文件名 进入下一层级
 function nextLayer(space_no){
    $(".zJMtAEb .file-name").each(function(){
        $(this).click(function(){
            let value = $(this).prev().prev().attr("data-folder");
            if(value=='false'){
                return
            }
            let path = $(this).prev().prev().attr("data-path");            
            //请求列表
             list(path,space_no,1000,1,'modtime',true);
              
        });
     });
 }


// <!--重命名-->
function rename(){
    let a = $(".zJMtAEb input[name='fileSelect']:checked").parent().parent().position().top;
    $('#file-rename-box').show();
    $('#file-rename-box').css('top',a);
    // <!--非选中的input 和 全选按钮禁止点击-->
    $(".zJMtAEb input[name='fileSelect']").not(':checked').attr('disabled','disabled');
    $("#s-selectAll").attr("disabled","disabled");
    // <!--要改的文件名字给input value-->
    let name = $(".zJMtAEb input[name='fileSelect']:checked").attr("data-name");
    $(".renameInput").val(name);
}
// <!--重命名div隐藏-->
function renameDivHide(){
    $(".renameInput").val('');
    $(".zJMtAEb input[name='fileSelect']").not(':checked').attr('disabled',false);
    $("#s-selectAll").attr("disabled",false);
    $("#file-rename-box").hide();
}
// <!--重命名点击事件-->
$("#renameBtn").click(function(){
    if($(".zJMtAEb input[name='fileSelect']:checked").size()==1){
        rename();
    }
});
// <!--重命名取消-->
$("#rename-cancel").click(function(){
    renameDivHide();
});
// <!--重命名确定-->
$("#rename-cfm").click(function(){
    let renameInputV= $(".renameInput").val();
    $(".zJMtAEb input[name='fileSelect']:checked").next().next().html(renameInputV);
    let dataId = $(".zJMtAEb input[name='fileSelect']:checked").attr("data-id");
    $.ajax({
        url:"/api/v1/store/rename",
        method:"POST",
        contentType: "application/json",
        data:JSON.stringify({
            "src":dataId,
            "dest":renameInputV
        }),
        success:function(res){
            console.log(res);
            if(res.Data=="ok"){
                //重命名成功，隐藏重命名组件
                renameDivHide();
            }
        }
    });
});

//-------------------------------------------------------------------------------------------------------------------------------------------------//



//上传文件
$("#upLoadFileBtn").click(function(){
    $("#upLoadFileIpt").unbind().change(function(){
        let localPath = document.getElementById("upLoadFileIpt").files[0].path;
        let space_no = '';
        let hashPath = method.getParamsUrl().path;
        let a = hashPath.split(":")[0];
        let b = hashPath.split(":")[1];   
        if(a == "myspace"){
            space_no = 0;
        }else if(a == "privite"){
            space_no = 1;
        }
        hashPath = b;
        if(b ==0){
            hashPath = "/";
        }

        console.log(localPath+";"+hashPath);

        $.ajax({
            url:"/api/v1/store/upload",
            method:"POST",
            contentType: "application/json",
            data:JSON.stringify({
                "filename":localPath,
                "dest_dir": hashPath,
                "interactive":true,
                "newversion" :false,
                "space_no":space_no,
                "is_encrypt":true
            }),
            success:function(res){
                console.log(res);
                if((res.code==0)&&(res.Data=='ok')){
                    method.firstInit();
                }
            }
        });
    });
});

//上传文件夹
$("#upLoadFolderBtn").click(function(){
    $("#upLoadFolderIpt").unbind().change(function(){
        let localPath = document.getElementById("upLoadFolderIpt").files[0].path;
        let space_no = '';
        let hashPath = method.getParamsUrl().path;
        let a = hashPath.split(":")[0];
        let b = hashPath.split(":")[1];   
        if(a == "myspace"){
            space_no = 0;
        }else if(a == "privite"){
            space_no = 1;
        }
        hashPath = b;
        if(b ==0){
            hashPath = "/";
        }

        console.log(localPath+";"+hashPath);

        $.ajax({
            url:"/api/v1/store/uploaddir",
            method:"POST",
            contentType: "application/json",
            data:JSON.stringify({
                "parent":localPath,
                "dest_dir": hashPath,
                "space_no":space_no,
                "is_encrypt":false
            }),
            success:function(res){
                console.log(res);
                if((res.code==0)&&(res.Data=='ok')){
                    method.firstInit();
                }
            }
        });

    });
});

//点击下载文件

//需要判断所有勾选的是文件还是文件夹，根据结果来执行哪个下载；
$("#downLoadBtn").click(function(){
    $("#downLoadIpt").unbind().change(function(){
        let localPath = document.getElementById("downLoadIpt").files[0].path;
        let selectedArr = $(".zJMtAEb input[name='fileSelect']:checked");
       $("#downLoadIpt1")[0].reset();
        $.each(selectedArr,function(index,obj){
            let filehash = $(obj).attr("data-hash");
            let filesize = Number($(obj).attr("data-size"));
            let filename = $(obj).attr("data-path");
            let space_no = Number($(obj).attr("data-spaceno"));
            let isFolder = $(obj).attr("data-folder");
            if(isFolder=='false'){
                $.ajax({
                    url:"/api/v1/store/download",
                    method:"POST",
                    contentType: "application/json",
                    data:JSON.stringify({
                        "filehash":filehash,
                        "filesize":filesize,
                        "filename":filename,
                        "space_no":space_no,
                        "dest_dir":localPath
                    }),
                    success:function(res){
                        console.log(res);
                        if(res.code==0&&res.Data=='ok'){
                            alert('Download success!');
                        }else{
                            alert(res.errmsg);
                        }
                    }
                });
            }else{
                $.ajax({
                    url:"/api/v1/store/downloaddir",
                    method:"POST",
                    contentType: "application/json",
                    data:JSON.stringify({
                        "parent":filename,
                        "space_no":space_no,
                        "dest_dir":localPath
                    }),
                    success:function(res){
                        console.log(res);
                        if(res.code==0&&res.Data=='ok'){
                            alert('Download success!');
                        }else{
                            alert(res.errmsg);
                        }
                    }
                });
            }
        });
    });
});


//删除
$("#deleteBtn").click(function(){
    let inputArr = $(".zJMtAEb input[name='fileSelect']:checked");
    let arr=[];
    //let json = {};
    for(i=0;i<inputArr.length;i++){
         let json = {};
        let id = $(inputArr[i]).attr("data-id");
        let path = $(inputArr[i]).attr("data-path");
        if(id){
            json.target = id;
            json.ispath = false;
        }else{
            json.target = path;
            json.ispath = true;
        }
        json['space_no']=parseInt($(inputArr[i]).attr("data-spaceNo"));
        json.recursion = true;
        arr.push(json);

    }
    console.log(arr);
    for(var i = 0;i<arr.length;i++){
        console.log(arr[i]);
        $.ajax({
            url:"/api/v1/store/remove",
            method:"POST",
            contentType: "application/json",
            data:JSON.stringify(arr[i]),
            success:function(res){
                console.log(res);
                if((res.code==0)&&(res.Data=="ok")){
                    method.firstInit();
                }else if(res.code!=0){
                    alert('Delete failed!');
                }
            }
        });
    }

   
});


//addFolder
function trigger(){
    let path = method.getParamsUrl().path;
    let space_no;
    //console.log(path);
    let a = path.split(":")[0];
    let b = path.split(":")[1];   
    if(a == "myspace"){
        space_no = 0;
    }else if(a == "privite"){
        space_no = 1;
    }
    path = b;
    if(b ==0){
        path = "/";
    }
    
    //console.log(path);
    //nofile隐藏
    $(".no-file-ab").hide();
    //js选插入一行
    let row =` <dd class="AuPKyz">
                    <div data-key="name" class="AuPKyz-li" style="width:60%;">
                        <input class="s-select-check" type="checkbox" name="fileSelect" data-name="" data-id="" data-path="${path}/new folder" data-hash="" data-folder=true  data-spaceNo="${space_no}">
                        <span class="file-icon "></span>
                        <span class="file-name appendFileName" title=""></span>
                        <span class="renameSpan">
                            <input class="renameInput" type="text" value="new folder" autofocus>
                            <span  class="rename-icon rename-cfm">
                                <svg t="1529648975355" class="icon"  viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4192" xmlns:xlink="http://www.w3.org/1999/xlink" width="25" height="25"><defs><style type="text/css"></style></defs><path d="M512 960C265.6 960 64 758.4 64 512S265.6 64 512 64s448 201.6 448 448-201.6 448-448 448z m0-832c-211.2 0-384 172.8-384 384s172.8 384 384 384 384-172.8 384-384-172.8-384-384-384z m-16 528c-8 0-16-3.2-22.4-9.6l-160-160c-12.8-12.8-12.8-32 0-44.8 12.8-12.8 32-12.8 44.8 0L496 579.2l217.6-217.6c12.8-12.8 32-12.8 44.8 0 12.8 12.8 12.8 32 0 44.8l-240 240C512 652.8 504 656 496 656z" p-id="4193" fill="#3b8cff"></path></svg>
                            </span>
                            <span  class="rename-icon rename-cancel">
                                <svg t="1529649149532" class="icon"  viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="5031" xmlns:xlink="http://www.w3.org/1999/xlink" width="23" height="23"><defs><style type="text/css"></style></defs><path d="M499.104 83.392c239.68 0 434.656 194.976 434.656 434.624 0 239.648-195.008 434.656-434.656 434.656S64.448 757.664 64.448 518.016 259.456 83.392 499.104 83.392m0 933.28c274.944 0 498.656-223.712 498.656-498.656S774.08 19.392 499.104 19.392C224.16 19.392 0.448 243.072 0.448 518.016s223.712 498.656 498.656 498.656" p-id="5032" fill="#3b8cff"></path><path d="M278.72 704.512h-0.032a31.968 31.968 0 1 0 45.248 45.28l186.496-186.56 186.528 186.56a31.968 31.968 0 1 0 45.248-45.248l-0.032-0.032-186.464-186.496 186.464-186.464 0.032-0.032a31.968 31.968 0 1 0-45.248-45.248l-186.528 186.496-186.496-186.496A31.968 31.968 0 1 0 278.72 331.52l186.464 186.464-186.464 186.496z" p-id="5033" fill="#3b8cff"></path></svg>
                            </span>
                        </span>
                    </div>
                    <div data-key="size" class="AuPKyz-li" style="width:16%;">
                        <span class="text">--</span>
                    </div>
                    <div data-key="time" class="AuPKyz-li" style="width:23%;">
                        <span class="text">--</span>
                    </div>
                </dd>`;
    $('.zJMtAEb').prepend(row);
   
    //取消点击事件，避免重复点击；
    $("#addFolderBtn").removeAttr("onclick");

    $(".rename-cancel").click(function(){
        $('.zJMtAEb dd').remove('.zJMtAEb dd:eq(0)');
        $("#addFolderBtn").attr("onclick",'trigger()');
   });
   $(".rename-cfm").click(function(){
       let name = $(".zJMtAEb dd:eq(0) .renameInput").val();
       $.ajax({
           url:"/api/v1/store/folder/add",
           method:'POST',
           contentType:'application/json',
           //创建目录  在 根目录下创建     abc  和 tmp 两个子目录
           data:JSON.stringify({
               "parent":path,
               "space_no":space_no,
               "folders":[name],
               "interactive":true
           }),
           success:function(res){
               if(res.Data == true){
                   $(".zJMtAEb dd:eq(0) .appendFileName").html(name);
                   $(".renameSpan").hide();
                   $(".zJMtAEb dd:eq(0) .s-select-check").attr("data-path",path+'/'+name);
                   $("#addFolderBtn").attr("onclick",'trigger()');
               }else{
                   alert('Please rename the folder!');
               }
           }
       });
   });

    // <!--单行选中增加类-->
    rowSelected();
    //点击文件名进入下一层
   // nextLayer(space_no);
}



//-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------//



// <!--全选反选-->
$("#s-selectAll").click(function(){
    let target = $(this).attr('data-check-target');
    $(target).prop('checked',$(this).prop('checked'));
    btngroupshow();
});


//设置隐私空间密码
$("#goInPriviteSpace1").click(function(){
    let setVal = $("#setPassword").val();
    let cfmVal = $("#confirmPassword").val();
    let reg = /^(?![0-9]+$)(?![a-z]+$)(?![A-Z]+$)(?![,\.#%'\+\*\-:;^_`]+$)[,\.#%'\+\*\-:;^_`0-9A-Za-z]{10,32}$/;
    if(!setVal||!cfmVal){
        alert("Password cannot be empty!");
        return;
    }else if(!reg.test(setVal)||!reg.test(cfmVal)){
        alert("Incorrect password setting rules!");
        return;
    }else if(setVal!=cfmVal){
        alert("The password entered twice is different!");
        return;
    }else{
        //密码正确了！
        method.setPrivtePsd(setVal,1);
    }
});



//输入密码进入空间
$("#goInPriviteSpace2").click(function(){
    let importPsd = $("#importPassword").val();
    let reg = /^(?![0-9]+$)(?![a-z]+$)(?![A-Z]+$)(?![,\.#%'\+\*\-:;^_`]+$)[,\.#%'\+\*\-:;^_`0-9A-Za-z]{10,32}$/;
    if(!importPsd){
        alert("enter your PIN");
    }else if(!reg.test(importPsd)){
        alert("Input password is not standard!");
    }else{
        //发送验证密码
        $.ajax({
            url:"/api/v1/space/verify",
            contentType:"application/json",
            method:'POST',
            data:JSON.stringify({
                "password":importPsd,
                "space_no":1
            }),
            success:function(res){
                console.log(res);
                if(res.code==0){
                    //已输入过密码；
                    sessionStorage.setItem("hadImport",true);
                    $("#priviteCondition").hide();
                    $("#diskDivAll").show();
                    list('/',1,1000,1,'modtime',true);  

                }else{
                    alert('Password incorrect!');
                }
            }
        });

    }
});




var public = {
    //秒转时间
    Date:function(s){
        Date.prototype.Format = function (fmt) { 
            var o = {
                "M+": this.getMonth() + 1, //月份 
                "d+": this.getDate(), //日 
                "h+": this.getHours(), //小时 
                "m+": this.getMinutes(), //分 
                "s+": this.getSeconds(), //秒 
                "q+": Math.floor((this.getMonth() + 3) / 3), //季度 
                "S": this.getMilliseconds() //毫秒 
            };
            if (/(y+)/.test(fmt)) fmt = fmt.replace(RegExp.$1, (this.getFullYear() + "").substr(4 - RegExp.$1.length));
            for (var k in o)
            if (new RegExp("(" + k + ")").test(fmt)) fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
            return fmt;
        }
        return new Date(s*1000).Format("yyyy-MM-dd hh:mm");
    },
    //四舍五入保留2位小数（若第二位小数为0，则保留一位小数）   //M转G/T
    unitConversion:function (num) {
        var result = parseFloat(num);
        if (isNaN(result)) {
            return false;
        }
        if(result<1024){
            return result+'M'
        }else if((result>=1024)&&(result<1024*1024)){
            return Math.round(result*100/1024)/100+'G';
        }else if((result>=1024*1024*1024)){
            return Math.round(result*100/1024*1024)/100+'T';
        }

    }
};




 //购买 确认订单
function addCar(id,no){
    if(sessionStorage.getItem("noPaidOrder")){
        let r=confirm("You have an unpaid order and the submission will overwrite the order.");
        if (r==true){
            $.ajax({
                url:"/api/v1/package/buy",
                contentType: "application/json",
                method:"POST",
                data:JSON.stringify({
                    "id":id,
                    "canceled":true,
                    "quanlity":no
                }),
                success:function(res){
                    console.log(res);
                    if(res.code==0){
                        //加购物车成功
                        $("#packageUl li").eq(1).addClass("packageOnLi").siblings().removeClass("packageOnLi");
                        $("#packageItem>div").hide().eq(1).show();
                        //刷新我的订单
                        packageMethod.orderAllInit();
                    }else{
                        alert("Operation failed, please re-operate ~");
                    }
                }
            });
        }
    }
}
//删除订单
function delOrder(id){
    $.ajax({
        url:"/api/v1/order/remove",
        contentType: "application/json",
        method:"POST",
        data:JSON.stringify({
            "order_id":id
        }),
        success:function(res){
            console.log(res);
            packageMethod.orderAllInit();
        }
    });
}
//付款
function pay(orderId){
    $.ajax({
        url:"/api/v1/order/pay",
        contentType: "application/json",
        method:"POST",
        data:JSON.stringify({
            "order_id":orderId
        }),
        success:function(res){
            if(res.code==0){
                alert('Payment success!');
            }else{
                alert('Payment failed, Please try again!');
            }
        }
    });
}


//我的传输入列表方法集
var transportMethod = {
    transportInit:function(){
        $.ajax({
            url:"/api/v1/store/progress",
            contentType: "application/json",
            method:"POST",
            data:JSON.stringify({
                "files":[]
            }),
            success:function(res){
                console.log(res);
               if((res.code==0)&&(JSON.stringify(res.Data)!="{}")){
                    let html = '';
                    $.each(res.Data,function(idx,obj){
                        //console.log(idx.split("\\")[idx.split("\\").length-1]);
                        html +=`<li class="tsList">
                                    <div class="tsList-l" title="${idx}">${idx.split("\\")[idx.split("\\").length-1]}</div>
                                    <div class="tsList-r">
                                        <div class="tsList-r-Bar">
                                            <div class="tsList-r-progressBar" style="width:${obj*100+'%'};">${Math.round(obj*100)+'%'}</div>
                                        </div>
                                    </div>
                                </li>`;
                    })
                    $("#tsMenu").html(html);  
                }else{
                    $("#tsMenu").html('no data!');
               }
            }
        });
    }
};


//我的套餐方法
var packageMethod = {
    //选项卡
    selectCard:function(){
        $("#packageUl li").click(function(){
            $(this).addClass("packageOnLi").siblings().removeClass("packageOnLi");
            $("#packageItem>div").hide().eq($(this).index()).show();
        });
    },
    //查询充值地址和余额 初始化
    queryBalanceInit:function (){
        $("#qrcode").html('');
        $.ajax({
            url:"/api/v1/order/recharge/address",
            "method":"GET",
            success:function(res){
                console.log(res);
                if(res.code==0){
                    $("#balanceNo").html(res.Data.balance/1000000);     //余额
                    $("#samosWalletAddress").val(res.Data.address);     //转账地址
                    $("#samosWalletCopyBtn").attr("data-clipboard-text",res.Data.address);      //复制转账地址属性；
                    jQuery("#qrcode").qrcode('samos//pay?address='+res.Data.address+'&amount=0&token=samos'); //二维码生成 
                }
            }
        });
    },
    
    
    // 所有订单初始化
    orderAllInit:function(){
        let expired = true;
        $.ajax({
            url:"/api/v1/order/all?expired="+expired,
            success:function(res){
                console.log(res);
                if(res.code==0){
                    if(res.Data.length>0){
                        let html = '';
                        $.each(res.Data,function(idx,obj){
                            if(!obj.paid){
                                sessionStorage.setItem("noPaidOrder",true);
                            }
                            html +=` <div class="order-list">
                                        <div class="order-list-head clearfix">
                                            <span data-locale="createdTime">创建日期：<span>${public.Date(obj.creation)}</span></span>
                                            <span data-locale="orderNumber">订单号：<span>${obj.id}</span></span>
                                            <div  class="order-list-del" onclick="delOrder(${obj.id})">
                                                ${obj.paid?'':'&times;'}
                                            </div>
                                        </div>
                                        <div class="order-list-body">
                                            <div class="order-list-name">
                                                <div class="order-list-t" data-locale="packageName">套餐名字</div>
                                                <div>${obj.package.name}</div>
                                            </div>
                                            <div class="order-list-inf">
                                                <div class="order-list-t" data-locale="packageInf">套餐信息</div>
                                                <div>
                                                    <span>price:${obj.package.price/1000000};</span>
                                                    <span>volume:${obj.package.volume};</span>
                                                    <span>netflow:${obj.package.netflow};</span>
                                                    <span>upNetflow:${obj.package.upNetflow};</span>
                                                    <span>downNetflow:${obj.package.downNetflow};</span>
                                                    <span>validDays:${obj.package.validDays};</span>
                                                </div>
                                            </div>
                                            <div class="order-list-inf">
                                                <div class="order-list-t" data-locale="allInf">总计信息</div>
                                                <div>
                                                    <span>price:${obj.totalAmount/1000000};</span>
                                                    <span>volume:${obj.volume};</span>
                                                    <span>netflow:${obj.netflow};</span>
                                                    <span>upNetflow:${obj.upNetflow};</span>
                                                    <span>downNetflow:${obj.downNetflow};</span>
                                                    <span>validDays:${obj.validDays};</span>
                                                    <span class="switchInfo ${obj.payTime?'':'hidden'}">
                                                        <span>pay time:${public.Date(obj.payTime)};</span>
                                                        <span>start time:${public.Date(obj.startTime)};</span>
                                                        <span>end time:${public.Date(obj.endTime)};</span>
                                                    </span>
                                                </div>
                                            </div>
                                            <div class="order-list-total">
                                                <div class="order-list-t" datalocale="totalAmount">总额</div>
                                                <div>${obj.totalAmount/1000000}</div>
                                            </div>
                                            <div class="order-list-quanlity">
                                                <div class="order-list-t" data-locale="amount">数量</div>
                                                <div>${obj.quanlity}</div>
                                            </div>
                                            <div class="order-list-pay">
                                                <div class="order-list-t" data-locale="buy">购买</div>
                                                <div class="${obj.paid?'':'toBuy'}" onclick="${obj.paid?'javascript:;':'pay(\''+obj.id+'\')'}">${obj.paid?'account paid':'buy'}</div>
                                            </div>
                                        </div>
                                    </div>`;
                        });
                        $("#orders").html(html);
                        $("#no-order").hide();
                        $("#orders").show();
                    }else{
                        $("#orders").hide();
                        $("#no-order").show();
                    }
                }else{
                    $("#no-order").show();
                }
            }
        });
    },

    //我的套餐 初始化
    packageInit:function (){
        //套餐
        let _this = this;
        $.ajax({
            url:"/api/v1/package/all",
            method:'GET',
            success:function(res){
                console.log(res);
                $('#pan-buy-menu').empty();
                if(!res.Data){return;}
                let html = '';
                $.each(res.Data,function(idx,obj){
                    html +=` <div class="pan-scheme">
                                <input id="${'ipt'+obj.id}" value="${obj.id}"  hidden/>
                                <div class="pan-scheme-item">
                                    <h2 id="${'name'+obj.id}">${obj.name}</h2>
                                    <p id="${'volume'+obj.id}">Space capacity:<span>${obj.volume}</span>G</p>
                                    <p id="${'netflow'+obj.id}">Network flow:<span>${obj.netflow}</span>G</p>
                                    <p id="${'upNetflow'+obj.id}">Upload:<span>${obj.upNetflow}</span>G</p>
                                    <p id="${'downNetflow'+obj.id}">Download:<span>${obj.downNetflow}</span>G</p>
                                    <p id="${'validDays'+obj.id}">Term of validity:<span>${obj.validDays}</span>days</p>
                                </div>
                                <div class="pan-select">select
                                    <select id="${'Select'+obj.id}">
                                        <option>1</option>
                                        <option>2</option>
                                        <option>3</option>
                                        <option>4</option>
                                        <option>5</option>
                                    </select>copys
                                </div>
                                <div class="pan-price">Amount payable：<span id="${'Price'+obj.id}" data-price ="${obj.price/1000000}">${obj.price/1000000}</span> samos</div>
                                <div id="${obj.id}" type="button" class="pan-buy-btn" onclick="addCar('${obj.id}',1)">buy</div>
                                </div>`;    
                });
                $('#pan-buy-menu').html(html);
                // <!--份数选择-->
                $('#pan-buy-menu select').change(function(){
                    let selectId = $(this).attr('id');
                    let id = selectId.substring(6);
                    $('#Price'+id).html($('#Price'+id).attr('data-price')*$(this).val());
                    $('#'+id).attr("onclick","addCar('"+id+"',"+$(this).val()+")");
                });
                //选项卡
                _this.selectCard();
                
            }

        });
        //查询充值地址和余额
        this.queryBalanceInit();
        //所有订单初始化
        this.orderAllInit();
    },
    //订单使用情况初始化
    amountInit:function(){
        $.ajax({
            url:"/api/v1/usage/amount",
            "method":"GET",
            success:function(res){
                console.log(res);
                if(res.code==0){
                   let obj = res.Data;
                    $("#volumeSpan").html(public.unitConversion(obj.volume));
                    $("#volumeUseSpan").html(public.unitConversion(obj.usageVolume));
                    $("#volumeUseBar").css('width',(obj.usageVolume/obj.volume)*100+'%');

                    $("#upNetFlowSpan").html(public.unitConversion(obj.upNetflow));
                    $("#upNetFlowUseSpan").html(public.unitConversion(obj.usageNetflow));
                    $("#upNetFlowUseBar").css('width',(obj.usageNetflow/obj.upNetflow)*100+'%');

                    $("#downNetFlowSpan").html(public.unitConversion(obj.downNetflow));
                    $("#downNetFlowUseSpan").html(public.unitConversion(obj.usageDownNetflow));
                    $("#downNetFlowUseBar").css('width',(obj.usageDownNetflow/obj.downNetflow)*100+'%');

                   $("#endTimeSpan").html(public.Date(obj.endTime));

                   $("#useCondition").show();
                }
            }
        });
    }
};




//转账刷新页面；
$("#initBalance").click(function(){
    packageMethod.queryBalanceInit();
});

//复制钱包地址
$("#samosWalletCopyBtn").click(function(){
    $('#samosWalletAddress').css('background','#3b8cff');
    var clipboard = new ClipboardJS('.samosWalletCopyBtn',{
        container: document.getElementById("myModal")
    });
    clipboard.on('success', function(e) {
        console.log(e.text);
    });
});
//刷新钱包金额
$("#refreshBtn").click(function(){
    packageMethod.queryBalanceInit();
});

//订单使用情况
packageMethod.amountInit();




















