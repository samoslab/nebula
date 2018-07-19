
function browserLang(){
    let lang = (navigator.languages && navigator.languages.length > 0) ? navigator.languages[0]
        : (navigator.language || navigator.userLanguage /* IE */ || 'en');
    lang = lang.toLowerCase();
    lang = lang.replace(/-/,"_"); // some browsers report language as en-US instead of en_US
    if (lang.length > 3) {
        lang = lang.substring(0, 3) + lang.substring(3).toUpperCase();
    }
    return lang;
}
function getLanguage(){
    let l = $.cookie("lang") ||browserLang();
    if(l==='zh_CN'){
        return l;
    }else{
        return 'en';
    }
}
function setLanguage(lang) {
    $.cookie("lang", lang, { expires: 30,path: '/' });
}

function loadProperties(name,type) {
    $.i18n.properties({
        name:name,
        path:'i18n/',
        mode:'map',
        language:type,
        callback:function(){
            $("[data-locale]").each(function(){
                if($(this).attr('title')){
                    $(this).attr('title',$.i18n.prop($(this).data("locale"))).html($.i18n.prop($(this).data("locale")));
                }else{
                    $(this).html($.i18n.prop($(this).data("locale")));
                }
            });
            //$("[data-locale-class]").each(function(){
            //    var prefix = $(this).data("locale-class")+'-'
            //    var cl =  $(this).attr("class").split(" ");
            //    for(var i=0;i<cl.length;i++){
            //        if(cl[i].indexOf(prefix)==0){
            //            $(this).removeClass(cl[i])
            //        }
            //    }
            //    $(this).addClass(prefix+getLanguage());
            //});
        }
    });
}

function showSwitch(lang){
    if(lang==='zh_CN'){
        $('#changeToCn').hide();
        $('#changeToEn').show();
    }else{
        $('#changeToCn').show();
        $('#changeToEn').hide();
    }
}
function changeLang(name, lang){
    //if(lang===getLanguage()){
    //    return
    //}
    setLanguage(lang);
    loadProperties(name, lang);
    showSwitch(lang);
}