//var refreshTransportTimer = null;
var method = {
     //判断是否注册过邮箱！
     login:function(){
        const {ipcRenderer} = require('electron'); 
        $.ajax({
            url:"/api/v1/service/status",
            method:"GET",
            success:function(res){
                if(!res.status){
                   //没有注册过！跳到注册页
                   console.log(res);
                   ipcRenderer.send('default');
                }
            }
        })
    },
    //获取hashName;
    getParamsUrl:function (){
        var hashName = location.hash.split("#")[1];//路由地址
        //console.log(location.hash);
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
                if(res.code==0){
                    $.cookie("privitePsd", 1, { expires: 365,path: '/' });
                    $("#priviteCondition").hide();
                    $("#diskDivAll").show();
                    that.priviteInit('/',space_no);
                }else{
                    alert(res.errmsg);
                }
            }
        });
    },
    //初始化哪个选项
    firstInit:function(){
        //clearInterval(refreshTransportTimer);
        let path = method.getParamsUrl().path;
        location.hash = path;
        console.log(path);
        let space_no = 0;
        let a = path.split(":")[0];

        //请求参做个记录
        $("#thisListAugs").attr("path",path.split(":")[1]);

         //左侧选项卡
        $("#frameAsideUl li").removeClass("active");
        if(a == "myspace"){
            space_no = 0;
            // //左侧选项卡
            $("#mySpace").parent().addClass("active");
            this.spaceSelect(path,space_no);
        }else if(a == "privite"){
            space_no = 1;
            //左侧选项卡
            $("#privteSpace").parent().addClass("active");
            this.spaceSelect(path,space_no);
        }else if(a=="transport"){
            this.transportInit();
        }else if(a=="package"){
            this.packageInit();
        }
        
    },
    //初始化我的空间或 隐藏空间
    spaceSelect:function (path,space_no){
        path = path.split(":")[1];
        if(path.length == 0){
            path = "/";
        }
        console.log(path);
        //按钮组隐藏
        $("#s-button-group").hide();
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
        //遍历按哪种类型的排序
        let sorttype = sessionStorage.getItem("viewtype");
        $("#viewTypeDrop>span[data-key='"+sorttype+"']").addClass('ugcOHtb').siblings().removeClass('ugcOHtb'); //给类名
        $("#listContent").html("");
        list(path,space_no,100,1,sorttype,true);  
        
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
            let sorttype = sessionStorage.getItem("viewtype");
            $("#viewTypeDrop>span[data-key='"+sorttype+"']").addClass('ugcOHtb').siblings().removeClass('ugcOHtb'); //给类名
            $("#listContent").html("");
            list(path,space_no,100,1,sorttype,true);  
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
       // refreshTransportTimer = setInterval(function(){
            transportMethod.transportInit();
        //},2000);
         
    },
    //我的套餐初始化
    packageInit:function(){
        $("#package").parent().addClass("active");
        $("#diskDiv").hide();
        $("#transportDiv").hide();
        $("#PackageDiv").show();
        //调用套餐初始的方法们；
        packageMethod.packageInit();
    },
    //增加文件夹确定事件
    addFolderCf:function(initPath,path,space_no){
        let name = $("#listContent dd:eq(0) .renameInput").val();
        let html = ``;

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
                console.log(res);
                if(res.Data == true){
                    $("#listContent dd:eq(0) .appendFileName").html(name).attr('href',`${'#'+initPath+'/'+name}`);
                    $("#listContent dd:eq(0) .oprate").attr('style','');
                    $(".renameSpan").hide();
                    $("#listContent dd:eq(0) .s-select-check").attr({
                        "data-path":`${path}${path=='/'?'':'/'}${name}`,
                        "data-name":name
                    });
                    $("#addFolderBtn").attr("onclick",'trigger()');
                }else{
                    alert('Please rename the folder!');
                }
            }
        });
    }

};

//排序选择
 $("#viewType").mouseenter(function(){
    $(".viewTypeDrop").show();
});
$("#viewTypeDrop>span").click(function(){
    $("#viewTypeDrop>span").removeClass('ugcOHtb');
    $(this).addClass('ugcOHtb');
    let key = $(this).attr("data-key");
    sessionStorage.setItem("viewtype", key);
    $("#viewTypeDrop").hide();
    method.firstInit();
});

//滚动加载
$("#listBox").scroll(function () { 
    let total = parseInt($("#thisListAugs").attr("total")),
        pagesize = parseInt($("#thisListAugs").attr("pagesize")),
        pagenum = parseInt($("#thisListAugs").attr("pagenum")),
        path = $("#thisListAugs").attr("path"),
        space_no = parseInt($("#thisListAugs").attr("space_no"));

    let scrollTop =Math.round($(this).scrollTop()),
        winHeight = $(this).height(), 
        contentHeight = $("#listContent").height();

   if((pagesize*pagenum<total)&&(scrollTop + winHeight>= contentHeight)){
        let sorttype = sessionStorage.getItem("viewtype");
        console.log(path,space_no,pagenum,sorttype);
        list(path,space_no,100,pagenum+1,sorttype,true,true); 
        console.log('到底了');
    }
});
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
function list(path,space_no,pagesize,pagenum,sorttype,ascorder,apd){
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
            if(res.code==400){
                const {ipcRenderer} = require('electron'); 
                      ipcRenderer.send('default');
                      return;
            }
            if(res.code!=0) {alert(res.errmsg);return;}
            //列表总数
            $("#listTotalNum").html(res.Data.total);
            //请求参做个记录
            $("#thisListAugs").attr({
                "path":path,
                "space_no":space_no,
                "pagesize":pagesize,
                "pagenum":pagenum,
                "sorttype":sorttype,
                "ascorder":ascorder,
                "total":res.Data.total
            });

            //插入列表内容；
            append(res,path,space_no,apd);
            //插入面包屑导航内容；
            let html = breadNav(path,space_no);
            $("#breadNav").html(html);

            //把面包屑最后一个的链接地址给到左侧导航栏，目的是左侧切换时回到原目录，记得以前位置；
            let hrefPath = $("#breadNav li:last a").attr("href");
            switch(space_no){
                case 0:
                    $("#mySpace").attr("href",hrefPath);
                    break;
                case 1:
                    $("#privteSpace").attr("href",hrefPath);
                    break;
            }
        }
    });
 }

 //列表中插入所有list
function append(res,path,space_no,apd){
    let html = '';
    if((res.Data.total<1)||(res.code!=0)){
        $('#listContent').html('');
        $(".no-file-ab").show();
        return false;
    }

    $(".no-file-ab").hide();

    //let typeArr=["epub", "otf", "woff", "gz", "doc", "eot", "pdf", "ps", "rtf", "cab", "xls", "ppt", "pptx", "xlsx", "docx", "7z", "bz2", "Z", "deb", "elf", "crx", "lz", "exe", "nes", "rar", "rpm", "swf", "sqlite", "tar", "ar", "xz", "zip", "amr", "m4a", "mid", "mp3", "ogg", "flac", "wav", "bmp", "gif", "jpg", "png", "tif", "psd", "jxr", "webp", "cr2", "ico", "mp4", "mpg", "mov", "webm", "flv", "m4v", "mkv", "wmv", "avi"];
   
    $.each(res.Data.files,function(index,obj){
        let a;
        if(obj.folder){
             a = 'folder';
        }else{
             a = obj.extension;
        }
        let k = obj.filesize; 
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
        
        html+=`<dd class="AuPKyz" onmouseenter = "oprateShow(this)" onmouseleave = "oprateHide(this)">
                    <div data-key="name" class="AuPKyz-li" style="width:44%;">
                        <input class="s-select-check" type="checkbox" name="fileSelect" data-name="${obj.filename}" data-id="${obj.id}" data-path="${path}${'/'+obj.filename}" data-hash="${obj.filehash}" data-size="${obj.filesize}" data-folder=${obj.folder} data-spaceNo="${space_no}">
                        <span class="file-icon my-file-${a}"></span>
                        <a class="file-name" title="${obj.filename}" href="${obj.folder?(no+path+'/'+obj.filename):'javascript:;'}">${obj.filename}</a>
                        <div class="oprate">
                            <a class="g-button g-btn-download" href="javascript:;" title="DownLoad"  onclick="gBtnDownLoad('${obj.id}','${obj.filehash}','${obj.filesize}','${path}${'/'+obj.filename}','${space_no}','${obj.folder}')">
                                <span>
                                    <svg t="1529565144872" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="6402" xmlns:xlink="http://www.w3.org/1999/xlink" width="16" height="16">
                                    <defs><style type="text/css"></style></defs><path d="M199.3216 552.704v284.3648h655.9232v-284.3648a31.3856 31.3856 0 0 1 62.8224 0v284.3648c0 34.6624-28.16 62.7712-62.8224 62.7712H199.3728c-34.7136 0-62.8224-28.1088-62.8224-62.7712v-284.3648a31.3856 31.3856 0 1 1 62.7712 0z" p-id="6403" fill="#3b8cff"></path><path d="M531.2 171.1104c17.408 0 31.4368 14.0288 31.4368 31.3856v487.2192H499.8144V202.496c0-17.3568 14.08-31.3856 31.3856-31.3856z" p-id="6404" fill="#3b8cff"></path><path d="M532.48 669.2352l128.3584-128.256a31.3856 31.3856 0 0 1 44.3904 44.3904L532.48 758.0672l-172.6976-172.6976a31.3856 31.3856 0 1 1 44.3904-44.3904l128.3072 128.256z" p-id="6405" fill="#3b8cff"></path>
                                    </svg>
                                </span>
                                <form class="h5-uploader-form">
                                    <input id="rowDownLoadIpt${obj.id}"  class="h5-uploader-form"  multiple="" webkitdirectory="" accept="*/*" type="file">
                                </form>
                            </a>
                            <a class="g-button g-btn-delete" href="javascript:;" title="Delete" onclick = "gBtnDelete('${obj.id}','${path}${'/'+obj.filename}','${obj.folder}','${space_no}',true)">
                                <span>
                                    <svg t="1529565511794" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7147" xmlns:xlink="http://www.w3.org/1999/xlink" width="16" height="16">
                                        <defs><style type="text/css"></style></defs><path d="M677.35552 204.8l0-61.44c0-33.91488-29.61408-61.44-66.1504-61.44l-198.49216 0c-36.57728 0-66.19136 27.52512-66.19136 61.44l0 61.44-264.64256 0 0 61.44 99.24608 0 0 614.4c0 33.95584 29.65504 61.44 66.1504 61.44l529.32608 0c36.57728 0 66.19136-27.48416 66.19136-61.44l0-614.4 99.24608 0 0-61.44L677.35552 204.8 677.35552 204.8zM412.71296 143.36l198.49216 0 0 61.44-198.49216 0L412.71296 143.36 412.71296 143.36zM776.6016 880.64 247.27552 880.64l0-614.4 529.32608 0L776.6016 880.64 776.6016 880.64zM346.5216 358.4l66.19136 0 0 430.08-66.19136 0L346.5216 358.4 346.5216 358.4zM478.86336 358.4l66.1504 0 0 430.08-66.1504 0L478.86336 358.4 478.86336 358.4zM611.20512 358.4l66.1504 0 0 430.08-66.1504 0L611.20512 358.4 611.20512 358.4zM611.20512 358.4" p-id="7148" fill="#3b8cff"></path>
                                    </svg>
                                </span>
                              
                            </a>
                            <a class="g-button g-btn-rename" href="javascript:;" title="Rename" onclick = "gBtnRename(this)">
                                <span>
                                    <svg t="1529565661279" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7381" xmlns:xlink="http://www.w3.org/1999/xlink" width="16" height="16">
                                        <defs><style type="text/css"></style></defs><path d="M949.2 63.44c-33.696-32.4-70.88-48.832-110.576-48.832-62.064 0-107.344 40.224-119.728 52.56C701.488 84.48 106.832 681.488 106.832 681.488c-3.888 3.92-6.72 8.8-8.176 14.176-13.424 49.728-80.592 270.416-81.264 272.624-3.456 11.296-0.384 23.616 7.92 31.936 5.984 5.952 13.936 9.168 22.08 9.168 3.2 0 6.432-0.496 9.6-1.52 2.288-0.752 229.248-74.384 266.608-85.568 4.912-1.472 9.424-4.16 13.072-7.792 23.6-23.408 578.224-573.824 615.04-611.968 38.08-39.392 56.992-80.384 56.256-121.872C1007.248 139.712 987.472 100.272 949.2 63.44zM285.472 744.288c-32.368-32.736-65.296-51.216-90.688-61.616 109.2-109.632 394.464-396 514.272-516.208 15.632 3.76 52.768 16.16 90.912 54.784 38.528 38.992 48.8 83.472 50.672 93.344-121.632 121.44-401.616 399.44-512 509.008C328.432 799.776 311.968 771.088 285.472 744.288zM152.736 735.104c16.96 4.512 52.272 17.6 88.32 54.08 27.76 28.096 40.8 58.992 46.656 77.856-43.008 13.872-137.152 46.48-196.976 65.84C108.48 874.336 138.4 783.2 152.736 735.104zM906.832 258.16c-1.264 1.312-3.36 3.424-5.856 5.968-9.776-25.28-26.928-57.728-56.624-87.792-30.336-30.704-61.12-48.8-85.792-59.52 2.096-2.096 3.728-3.728 4.368-4.368 3.536-3.504 35.664-34.32 75.696-34.32 23.056 0 45.68 10.544 67.296 31.344 25.632 24.672 38.848 49.024 39.28 72.368C945.616 205.696 932.72 231.36 906.832 258.16z" fill="#3b8cff" p-id="7382"></path>
                                    </svg>
                                </span>
                            </a>
                        </div>
                    </div>
                    <div data-key="size" class="AuPKyz-li" style="width:16%;">
                        <span class="text">${k}</span>
                    </div>
                    <div data-key="type" class="AuPKyz-li" style="width:16%;">
                        <span class="text">${(obj.folder==true)?'folder':obj.filetype}</span>
                    </div>
                    <div data-key="time" class="AuPKyz-li" style="width:23%;">
                        <span class="text">${public.Date(obj.modtime)}</span>
                    </div>
                </dd>`;
    });
    if(apd){
        $('#listContent').append(html);
    }else{
        $('#listContent').html(html);
    }
    //$('#listContent').append(html);
   // $('#listContent').html(html);
    //$('#listContent').html(`${html}`);

    // <!--单行选中增加类-->
    rowSelected(); 
}


//<!--列表是否选中是否显示按钮组 全选中的情况下 全选按钮也选中  如若选中数量大于1个禁止重命名按按钮点击事件-->
function btngroupshow(){
    if($("#listContent input[name='fileSelect']:checked").size()!=0){
        $("#s-button-group").show();
    }else{
        $("#s-button-group").hide();
    }
    if($("#listContent input[name='fileSelect']:checked").size()==$("#listContent input[name='fileSelect']").length){
        $("#s-selectAll").prop('checked','checked');
        $(".AuPKyz").addClass("activeSelect");
    }else{
        $("#s-selectAll").prop('checked',false);
    }
    if($("#listContent input[name='fileSelect']:checked").size()!=1){
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
    $("#listContent input[name='fileSelect']").each(function(){
        $(this).change(function(){
         //$(".s-select-check").change(function(){
             if($(this).is(":checked")){
                 //console.log($(this).attr("data-name"));查 选中行的文件夹名字
                 $(this).parents(".AuPKyz").addClass("activeSelect");
             }else{
                 $(this).parents(".AuPKyz").removeClass("activeSelect");
             }
             btngroupshow();
         });
     });
 }

 //面包屑导航 得到层级 该有的代码；
 function breadNav(path,space_no){
     let a = "";
     let lan = getLanguage();
    if(space_no==0){
        a = "#myspace:"
    }else if(space_no==1){
        a = "#privite:"
    }
    let pathArr1 = path.slice(1).split('/');
    let pathArr2 = [...pathArr1];
    let liHtml = '';
    let tmpArr = [];
    for(let i = 1;i<=pathArr1.length;i++){
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
                        <a href="${a}" title="all files"  data-locale="allFileOther">${(lan=='en')?'all files':'全部文件'}</a>
                    </li>`;
    allFileHtml+=liHtml;
    return allFileHtml;
 }

/*-------------------------------------------------重命名-----------------------------------------------------------------------------*/
var rename = {
    // <!--重命名条框显示-->
    rename:function (a,name){
        $('#file-rename-box').css('top',a);
        $('#file-rename-box').show();
        $("#file-rename-box .renameInput").val(name);

        // <!--非选中的input 和 全选按钮禁止点击-->
        $("#listContent input[name='fileSelect']").not(':checked').attr('disabled','disabled');
        $("#s-selectAll").attr("disabled","disabled");
    },
    // <!--重命名div隐藏-->
    renameDivHide:function (){
        $("#file-rename-box .renameInput").val('');
        $("#listContent input[name='fileSelect']").not(':checked').attr('disabled',false);
        $("#s-selectAll").attr("disabled",false);
        $("#file-rename-box").hide();
    },
    //重命名确定
    renameCf:function(){
        let renameInputV= $("#file-rename-box .renameInput").val();
        $("#listContent input[name='fileSelect']:checked").siblings(".file-name").html(renameInputV).attr("title",renameInputV);
        $("#listContent input[name='fileSelect']:checked").attr("data-name",renameInputV);
        //let dataId = $("#listContent input[name='fileSelect']:checked").attr("data-id");
        let dataPath = $("#listContent input[name='fileSelect']:checked").attr("data-path");
        let spaceNo =  Number($("#listContent input[name='fileSelect']:checked").attr("data-spaceno"));
        console.log(dataPath);

        let b = dataPath.lastIndexOf('/');
        let c = dataPath.substring(0,b+1);

        $.ajax({
            url:"/api/v1/store/rename",
            method:"POST",
            contentType: "application/json",
            //data:JSON.stringify(json),
            data:JSON.stringify({
                "src":dataPath,
                "dest":renameInputV,
                "space_no":spaceNo,
                "ispath":true
            }),
            success:function(res){
                console.log(res);
                if(res.code==0){
                    //重命名成功，隐藏重命名组件
                    rename.renameDivHide();
                      //重命名后给input赋值；
                    $("#listContent input[name='fileSelect']:checked").attr("data-path",c+renameInputV);
                }else if(res.code!=0){
                    alert(res.errmsg);
                }
            }
        });
    },
};

// <!--重命名点击事件-->
$("#renameBtn").click(function(){
    if($("#listContent input[name='fileSelect']:checked").size()==1){
        //选中行到顶部的距离
        let a = $("#listContent input[name='fileSelect']:checked").parents(".AuPKyz").position().top;
         // <!--要改的文件名字给input value-->
        let name = $("#listContent input[name='fileSelect']:checked").attr("data-name");
        rename.rename(a,name);
    }

});
 //<!--重命名取消-->
 $("#rename-cancel").click(function(){
    rename.renameDivHide();
});
//<!--重命名确定-->
$("#rename-cfm").click(function(){
    rename.renameCf();
});
//回车事件
$('.renameInput').keyup(function(event){
    if(event.keyCode ==13){
        rename.renameCf();
    }
});
/*---------------------------------------------------oprate 操作相关------------------------------------------------------------------------------------*/
//oprate显示隐藏
function oprateShow(x){
    //行选中数
    if($("#listContent input[name='fileSelect']:checked").size()<=1){
        $(x).find("div:first-child .oprate").show();
    }
}
function oprateHide(x){
    $(x).find("div:first-child .oprate").hide();
}
//行内重命名操作；
function gBtnRename(x){
    //位置
    let a = $(x).parents(".AuPKyz").position().top;
    // <!--要改的文件名字给input value-->
    let name = $(x).parent().siblings(".s-select-check").attr("data-name");

    $(".s-select-check").prop("checked",false);
    $("#listBox>dd").removeClass("activeSelect");
    $(x).parent().siblings(".s-select-check").prop("checked",true);
    rename.rename(a,name);
}


//// //单行下载按钮点击事件下载
function gBtnDownLoad(id,filehash,filesize,filename,space_no,isFolder){
    $("#rowDownLoadIpt"+id).change(function(){
        //console.log(id,filehash,filesize,filename,space_no,isFolder);
        let size = Number(filesize);
        let space = Number(space_no);
        let newId = "rowDownLoadIpt"+id;
        let localPath = document.getElementById(newId).files[0].path;
        console.log(localPath);
        if(isFolder=='false'){
            $.ajax({
                url:"/api/v1/store/download",
                method:"POST",
                contentType: "application/json",
                data:JSON.stringify({
                    "filehash":filehash,
                    "filesize":size,
                    "filename":filename,
                    "space_no":space,
                    "dest_dir":localPath
                }),
                success:function(res){
                    console.log(res);
                    if(res.code==0){
                        $("#updownGif").show();
                        // alert('Download success!');
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
                    "space_no":space,
                    "dest_dir":localPath
                }),
                success:function(res){
                    console.log(res);
                    if(res.code==0){
                        $("#updownGif").show();
                        // alert('Download success!');
                    }else{
                        alert(res.errmsg);
                    }
                }
            });
        }
    
    });
}

// //单行删除按钮点击事件删除
function gBtnDelete(id,path,ispath,space_no,recursion){
    //console.log(id+','+path+','+ispath+','+space_no+','+recursion);
    let a =confirm("Confirm to delete the selected file?");
    if(a==true){
        let renameInputV = $(".renameInput").val();
        console.log(path+renameInputV);
        let json = {};
        if(id){
            json.target = id;
            json.ispath = false;
        }else{
            json.target = path+renameInputV;
            json.ispath = true;
        }
        json['space_no']=parseInt(space_no);
        json.recursion = recursion;
        console.log(json);
        $.ajax({
            url:"/api/v1/store/remove",
            method:"POST",
            contentType: "application/json",
            data:JSON.stringify(json),
            success:function(res){
                console.log(res);
                if(res.code==0){
                    method.firstInit();
                    $("#file-rename-box").hide(); //重命名输入框显示时，又点击了删除按钮；
                }else{
                    alert('Delete failed!');
                }
            }
        });
    }

    
   
}



//---------------------------------------------------上传文件----------------------------------------------------------------------------------------------//



//上传文件
$("#upLoadFileBtn").click(function(){
    $("#upLoadFileIpt").unbind().change(function(){
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
        //选择的文件
        let local = document.getElementById("upLoadFileIpt").files;
        for(let i=0;i<local.length;i++){
            //每个选择文件的路径
            let localPath = local[i].path;
            $.ajax({
                url:"/api/v1/task/upload",
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
                    $("#updownGif").show();
                    // if(res.code==0){
                    //     method.firstInit();
                    // }
                }
            });
        }
    });
});

//上传文件夹
$("#upLoadFolderBtn").click(function(){
    $("#upLoadFolderIpt").unbind().change(function(){
        let localPath = document.getElementById("upLoadFolderIpt").files[0].path;
        if(!localPath)return;
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
            url:"/api/v1/task/uploaddir",
            method:"POST",
            contentType: "application/json",
            data:JSON.stringify({
                "parent":localPath,
                "dest_dir": hashPath,
                "interactive":true,
                "space_no":space_no,
                "is_encrypt":false
            }),
            success:function(res){
                console.log(res);
                $("#updownGif").show();
                // if(res.code==0){
                //     method.firstInit();
                // }
            }
        });

    });
});

//顶部点击下载文件
//需要判断所有勾选的是文件还是文件夹，根据结果来执行哪个下载；
$("#downLoadBtn").click(function(){
    $("#downLoadIpt").unbind().change(function(){
        let localPath = document.getElementById("downLoadIpt").files[0].path;
        let selectedArr = $("#listContent input[name='fileSelect']:checked");
       $("#downLoadIpt1")[0].reset();
        $.each(selectedArr,function(index,obj){
            let filehash = $(obj).attr("data-hash");
            let filesize = Number($(obj).attr("data-size"));
            let filename = $(obj).attr("data-path");
            let space_no = Number($(obj).attr("data-spaceno"));
            let isFolder = $(obj).attr("data-folder");
            if(isFolder=='false'){
                $.ajax({
                    url:"/api/v1/task/download",
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
                        if(res.code==0){
                            $("#updownGif").show();
                            //alert('Download success!');
                        }else{
                            alert(res.errmsg);
                        }
                    }
                });
            }else{
                $.ajax({
                    url:"/api/v1/task/downloaddir",
                    method:"POST",
                    contentType: "application/json",
                    data:JSON.stringify({
                        "parent":filename,
                        "space_no":space_no,
                        "dest_dir":localPath
                    }),
                    success:function(res){
                        console.log(res);
                        if(res.code==0){
                            $("#updownGif").show();
                           // alert('Download success!');
                        }else{
                            alert(res.errmsg);
                        }
                    }
                });
            }
        });
    });
});




//顶部删除按钮；
$("#deleteBtn").click(function(){
    let a =confirm("Confirm to delete the selected file?");
    if(a==true){
        let inputArr = $("#listContent input[name='fileSelect']:checked");
        let arr=[];
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
                    if(res.code==0){
                        method.firstInit();
                        $("#file-rename-box").hide(); //重命名输入框显示时，又点击了删除按钮；
                    }else{
                        alert('Delete failed!');
                    }
                }
            });
        } 
    }
});



//addFolder
function trigger(){
    let path = method.getParamsUrl().path;
    let initPath = path;
    let space_no;
    console.log(path);
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
    
    //nofile隐藏
    $(".no-file-ab").hide();
    //js选插入一行
    let row =` <dd class="AuPKyz">
                    <div data-key="name" class="AuPKyz-li" style="width:44%;">
                        <input class="s-select-check" type="checkbox" name="fileSelect" data-name="" data-id="" data-path="${path}${path=='/'?'':'/'}new folder" data-hash="" data-folder=true  data-spaceNo="${space_no}">
                        <span class="file-icon "></span>
                        <a class="file-name appendFileName" title="" href=""></a>
                        <span class="renameSpan">
                            <input class="renameInput" type="text" value="new folder" autofocus>
                            <span  class="rename-icon rename-cfm">
                                <svg t="1529648975355" class="icon"  viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4192" xmlns:xlink="http://www.w3.org/1999/xlink" width="25" height="25"><defs><style type="text/css"></style></defs><path d="M512 960C265.6 960 64 758.4 64 512S265.6 64 512 64s448 201.6 448 448-201.6 448-448 448z m0-832c-211.2 0-384 172.8-384 384s172.8 384 384 384 384-172.8 384-384-172.8-384-384-384z m-16 528c-8 0-16-3.2-22.4-9.6l-160-160c-12.8-12.8-12.8-32 0-44.8 12.8-12.8 32-12.8 44.8 0L496 579.2l217.6-217.6c12.8-12.8 32-12.8 44.8 0 12.8 12.8 12.8 32 0 44.8l-240 240C512 652.8 504 656 496 656z" p-id="4193" fill="#3b8cff"></path></svg>
                            </span>
                            <span  class="rename-icon rename-cancel">
                                <svg t="1529649149532" class="icon"  viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="5031" xmlns:xlink="http://www.w3.org/1999/xlink" width="23" height="23"><defs><style type="text/css"></style></defs><path d="M499.104 83.392c239.68 0 434.656 194.976 434.656 434.624 0 239.648-195.008 434.656-434.656 434.656S64.448 757.664 64.448 518.016 259.456 83.392 499.104 83.392m0 933.28c274.944 0 498.656-223.712 498.656-498.656S774.08 19.392 499.104 19.392C224.16 19.392 0.448 243.072 0.448 518.016s223.712 498.656 498.656 498.656" p-id="5032" fill="#3b8cff"></path><path d="M278.72 704.512h-0.032a31.968 31.968 0 1 0 45.248 45.28l186.496-186.56 186.528 186.56a31.968 31.968 0 1 0 45.248-45.248l-0.032-0.032-186.464-186.496 186.464-186.464 0.032-0.032a31.968 31.968 0 1 0-45.248-45.248l-186.528 186.496-186.496-186.496A31.968 31.968 0 1 0 278.72 331.52l186.464 186.464-186.464 186.496z" p-id="5033" fill="#3b8cff"></path></svg>
                            </span>
                        </span>
                        <div class="oprate" style="display:none;">
                            <a class="g-button g-btn-delete" href="javascript:;" title="Delete" onclick = "gBtnDelete('','${path}${path=='/'?'':'/'}',true,'${space_no}',true)">
                                <span>
                                    <svg t="1529565511794" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7147" xmlns:xlink="http://www.w3.org/1999/xlink" width="16" height="16">
                                        <defs><style type="text/css"></style></defs><path d="M677.35552 204.8l0-61.44c0-33.91488-29.61408-61.44-66.1504-61.44l-198.49216 0c-36.57728 0-66.19136 27.52512-66.19136 61.44l0 61.44-264.64256 0 0 61.44 99.24608 0 0 614.4c0 33.95584 29.65504 61.44 66.1504 61.44l529.32608 0c36.57728 0 66.19136-27.48416 66.19136-61.44l0-614.4 99.24608 0 0-61.44L677.35552 204.8 677.35552 204.8zM412.71296 143.36l198.49216 0 0 61.44-198.49216 0L412.71296 143.36 412.71296 143.36zM776.6016 880.64 247.27552 880.64l0-614.4 529.32608 0L776.6016 880.64 776.6016 880.64zM346.5216 358.4l66.19136 0 0 430.08-66.19136 0L346.5216 358.4 346.5216 358.4zM478.86336 358.4l66.1504 0 0 430.08-66.1504 0L478.86336 358.4 478.86336 358.4zM611.20512 358.4l66.1504 0 0 430.08-66.1504 0L611.20512 358.4 611.20512 358.4zM611.20512 358.4" p-id="7148" fill="#3b8cff"></path>
                                    </svg>
                                </span>
                            </a>
                            <a class="g-button g-btn-rename" href="javascript:;" title="Rename">
                                <span>
                                    <svg t="1529565661279" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7381" xmlns:xlink="http://www.w3.org/1999/xlink" width="16" height="16">
                                        <defs><style type="text/css"></style></defs><path d="M949.2 63.44c-33.696-32.4-70.88-48.832-110.576-48.832-62.064 0-107.344 40.224-119.728 52.56C701.488 84.48 106.832 681.488 106.832 681.488c-3.888 3.92-6.72 8.8-8.176 14.176-13.424 49.728-80.592 270.416-81.264 272.624-3.456 11.296-0.384 23.616 7.92 31.936 5.984 5.952 13.936 9.168 22.08 9.168 3.2 0 6.432-0.496 9.6-1.52 2.288-0.752 229.248-74.384 266.608-85.568 4.912-1.472 9.424-4.16 13.072-7.792 23.6-23.408 578.224-573.824 615.04-611.968 38.08-39.392 56.992-80.384 56.256-121.872C1007.248 139.712 987.472 100.272 949.2 63.44zM285.472 744.288c-32.368-32.736-65.296-51.216-90.688-61.616 109.2-109.632 394.464-396 514.272-516.208 15.632 3.76 52.768 16.16 90.912 54.784 38.528 38.992 48.8 83.472 50.672 93.344-121.632 121.44-401.616 399.44-512 509.008C328.432 799.776 311.968 771.088 285.472 744.288zM152.736 735.104c16.96 4.512 52.272 17.6 88.32 54.08 27.76 28.096 40.8 58.992 46.656 77.856-43.008 13.872-137.152 46.48-196.976 65.84C108.48 874.336 138.4 783.2 152.736 735.104zM906.832 258.16c-1.264 1.312-3.36 3.424-5.856 5.968-9.776-25.28-26.928-57.728-56.624-87.792-30.336-30.704-61.12-48.8-85.792-59.52 2.096-2.096 3.728-3.728 4.368-4.368 3.536-3.504 35.664-34.32 75.696-34.32 23.056 0 45.68 10.544 67.296 31.344 25.632 24.672 38.848 49.024 39.28 72.368C945.616 205.696 932.72 231.36 906.832 258.16z" fill="#3b8cff" p-id="7382"></path>
                                    </svg>
                                </span>
                            </a>
                        </div>
                    </div>

                    <div data-key="size" class="AuPKyz-li" style="width:16%;">
                        <span class="text">--</span>
                    </div>
                    <div data-key="type" class="AuPKyz-li" style="width:16%;">
                        <span class="text">folder</span>
                    </div>
                    <div data-key="time" class="AuPKyz-li" style="width:23%;">
                        <span class="text">--</span>
                    </div>
                </dd>`;
    $('#listContent').prepend(row);
   
    //取消点击事件，避免重复点击；
    $("#addFolderBtn").removeAttr("onclick");

    $(".rename-cancel").click(function(){
        $('#listContent dd').remove('#listContent dd:eq(0)');
        $("#addFolderBtn").attr("onclick",'trigger()');
   });
   $(".rename-cfm").click(function(){
       method.addFolderCf(initPath,path,space_no);
   });
   $('.renameInput').keyup(function(event){
        if(event.keyCode ==13){
            method.addFolderCf(initPath,path,space_no);
        }
    });

    // <!--单行选中增加类-->
    rowSelected();
}



//-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------//



// <!--全选反选-->
$("#s-selectAll").click(function(){
    let target = $(this).attr('data-check-target');
    $(target).prop('checked',$(this).prop('checked'));
    if(!$(this).prop('checked')){
        $(target).parents(".AuPKyz").removeClass("activeSelect");
    }
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
                    list('/',1,100,1,'modtime',true);  

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

    },
    //求窗口尺寸大小
    wh:function(){
        let fmh = $(".frame-main").height();
        $("#listBox").css('height',(fmh-106)+'px');
        $("#tsMenu").css('height',(fmh-106)+'px');
    }
};




 //购买 确认订单
function addCar(id,no){
    if(sessionStorage.getItem("noPaidOrder")=='true'){
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
    }else{
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
            console.log(res);
            if(res.code==0){
                alert('Payment success!');
            }else{
                alert(res.errmsg);
            }
        }
    });
}


//我的传输入列表方法集  ${idx.split("\\")[idx.split("\\").length-1]}
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
               if((res.code==0)&&(JSON.stringify(res.Data.progress)!="{}")){
                    let html = '';
                    $.each(res.Data.progress,function(idx,obj){
                    
                        html +=`<li class="tsList">
                                    <div class="tsList-l" title="${idx}">
                                        <span>${obj.type}</span>
                                        <span>${(idx.split('@')[0]=='0')?"My Space":"Privacy Space"}</span>
                                        <span> Path is：${idx.split('@')[1]}</span>
                                    </div>
                                    <div class="tsList-r">
                                        <div class="tsList-r-Bar">
                                            <div class="tsList-r-progressBar" data-name="${idx}" style="width:${obj.rate*100+'%'};">${Math.round(obj.rate*100)+'%'}</div>
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
    //查询转账情况
    inquireBalance:function(){
        let code = $("#samosWalletAddress").val();
        const {ipcRenderer} = require('electron'); 
        ipcRenderer.send('explorer',code);
    },
    
    // 所有订单初始化
    orderAllInit:function(){
        let expired = true;
        let lan = getLanguage();
        console.log(lan);
        $.ajax({
            url:"/api/v1/order/all?expired="+expired,
            success:function(res){
                console.log(res);
                if(res.code==0){
                    if(res.Data.length>0){
                        let html = '';
                        $.each(res.Data,function(idx,obj){
                            //有未付款的订单；
                            if(!obj.paid){
                                sessionStorage.setItem("noPaidOrder",true);
                            }else{
                                sessionStorage.setItem("noPaidOrder",false);
                            }
                            html +=` <div class="order-list">
                                        <div class="order-list-head clearfix">
                                            <span><span data-locale="createdTime">${(lan=='en')?'createdTime:':'创建时间:'}</span><span>${public.Date(obj.creation)}</span></span>
                                            <span><span data-locale="orderNumber">${(lan=='en')?'Order Number:':'订单号:'}</span><span>${obj.id}</span></span>
                                            <div  class="order-list-del" onclick="delOrder('${obj.id}')">
                                                ${obj.paid?'':'&times;'}
                                            </div>
                                        </div>
                                        <div class="order-list-body">
                                            <div class="order-list-name">
                                                <div class="order-list-t" data-locale="packageName">${(lan=='en')?'Package Name':'套餐名称'}</div>
                                                <div>${obj.package.name}</div>
                                            </div>
                                            <div class="order-list-inf">
                                                <div class="order-list-t" data-locale="packageInf">${(lan=='en')?'Package information':'套餐信息'}</div>
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
                                                <div class="order-list-t" data-locale="allInf">${(lan=='en')?'Total information':'总计'}</div>
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
                                                <div class="order-list-t" data-locale="totalAmount">${(lan=='en')?'Total Amount':'总额'}</div>
                                                <div>${obj.totalAmount/1000000}</div>
                                            </div>
                                            <div class="order-list-quanlity">
                                                <div class="order-list-t" data-locale="amount">${(lan=='en')?'Amount':'数量'}</div>
                                                <div>${obj.quanlity}</div>
                                            </div>
                                            <div class="order-list-pay">
                                                <div class="order-list-t" data-locale="buy">${(lan=='en')?'Buy':'购买'}</div>
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
//转账明细查询
$("#inquireBalance").click(function(){
    packageMethod.inquireBalance();
});
//刷新钱包金额
$("#refreshBtn").click(function(){
    packageMethod.queryBalanceInit();
});

//订单使用情况
packageMethod.amountInit();


//建立websocket
var wsServer = "ws://127.0.0.1:7799/message"; //服务器地址
var websocket = new WebSocket(wsServer); //创建WebSocket对象

console.log(websocket.readyState);//查看websocket当前状态
websocket.onopen = function (evt) {
    console.log(evt);//已经建立连接
};
websocket.onclose = function (evt) {
    console.log(evt);
    console.log('colose');//已经关闭连接
    if(evt){
        var r=confirm("Sorry,An error in the system requires a reboot!");
        if (r==true){
            const {ipcRenderer} = require('electron'); 
            ipcRenderer.send('close');
        }else{
           alert("You pressed Cancel! Please restart!");
        }
    }
};
websocket.onmessage = function (evt) {
    //收到服务器消息，使用evt.data提取
    console.log(evt.data);
    if(!evt.data)return;
    let data = JSON.parse(evt.data);
    if(((data.type=="DownloadProgress")||(data.type=="UploadProgress"))&&(data.progress!=1)){
        let key = data.key;
        $(".tsList-r-progressBar[data-name='"+key+"']").css("width",data.progress*100+'%').html(data.progress*100+'%');
        $("#updownGif").show();
    //}else if((data.type=="DownloadFile")||(data.type=="UploadFile")||(data.progress==1)){
    }else if((data.type=="DownloadFile")||(data.type=="UploadFile")){
        let key = data.key;
        $(".tsList-r-progressBar[data-name='"+key+"']").css("width",'100%').html('100%');
        method.firstInit();
        $("#updownGif").hide();
    }
   

};
websocket.onerror = function (evt) {
//产生异常
    console.log(evt);
    if(evt){
        var r=confirm("Sorry,An error in the system requires a reboot!");
        if (r==true){
            const {ipcRenderer} = require('electron'); 
            ipcRenderer.send('close');
        }else{
            alert("You pressed Cancel! Please restart!");
        }
    }
}; 










