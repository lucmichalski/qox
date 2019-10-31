"use strict";var _typeof="function"==typeof Symbol&&"symbol"==typeof Symbol.iterator?function(e){return typeof e}:function(e){return e&&"function"==typeof Symbol&&e.constructor===Symbol&&e!==Symbol.prototype?"symbol":typeof e};!function(e){"function"==typeof define&&define.amd?define(["jquery"],e):e("object"===("undefined"==typeof exports?"undefined":_typeof(exports))?require("jquery"):jQuery)}(function(e){function t(o,i){this.$element=e(o),this.options=e.extend({},t.DEFAULTS,e.isPlainObject(i)&&i),this.init()}var o="qor.help",i="enable."+o,n="disable."+o,r="click."+o,s="keyup."+o,a="change."+o,l=".qor-help__index";return t.prototype={constructor:t,init:function(){this.bind(),this.$newLink=e(".qor-slideout .qor-slideout__opennew"),this.$newLinkHref=this.$newLink.attr("href")},bind:function(){this.$element.on(r,".qor-help__lists [data-inline-url]",this.loadDoc.bind(this)).on(s,".qor-help__search",this.searchKeyup.bind(this)).on(r,".qor-help__search-button",this.search.bind(this)).on(a,".qor-help__search-category",this.search.bind(this)).on(r,".qor-pagination a",this.pagination.bind(this))},unbind:function(){this.$element.off(r,".qor-help__lists [data-inline-url]",this.loadDoc.bind(this)).off(s,".qor-help__search",this.searchKeyup.bind(this)).off(r,".qor-help__search-button",this.search.bind(this)).off(a,".qor-help__search-category",this.search.bind(this)).off(r,".qor-pagination a",this.pagination.bind(this))},pagination:function(o){var i=e(o.target),n=i.prop("href"),r=e(t.TEMPLATE_LOADING),s=this.$element.find(l);return e.ajax(n,{method:"GET",dataType:"html",beforeSend:function(){s.hide(),r.prependTo(s),window.componentHandler.upgradeElement(r.children()[0])},success:function(t){s.show().html(e(t).find(l))},error:function(e){r.remove(),window.alert(e.responseText)}}),!1},searchKeyup:function(e){13==e.keyCode&&this.searchAction()},search:function(){this.searchAction()},searchAction:function(){var o=e(".qor-help__search-category"),i=e(".qor-help__search"),n=e(".qor-help__body"),r=e(t.TEMPLATE_LOADING),s=[i.data().helpFilterUrl,"?",o.prop("name"),"=",o.val(),"&",i.prop("name"),"=",i.val()].join("");e.ajax(s,{method:"GET",dataType:"html",processData:!1,contentType:!1,beforeSend:function(){n.hide().after(r),window.componentHandler.upgradeElement(r.children()[0])},success:function(t){e(".qor-slideout__title").show(),e(".qor-slideout__show_title").remove(),n.html(e(t).find(".qor-help__body").html()).show(),r.remove()},error:function(e,t,o){n.show(),r.remove(),window.alert([t,o].join(": "))}})},loadDoc:function(o){var i,n=e(o.target),r=e(".qor-help__index"),s=e(t.TEMPLATE_LOADING),a=e(".qor-help__body"),l={},h=n.data().inlineUrl,c=this.$newLink,d=this.$newLinkHref;if(e(".qor-slideout").is(":visible"))return r.hide(),e.ajax(h,{method:"GET",dataType:"html",processData:!1,contentType:!1,beforeSend:function(){s.appendTo(a).trigger("enable")},success:function(o){e(o).find(".qor-page__show").appendTo(a).addClass("qor-doc__preview"),e(".qor-slideout__show_title").remove(),e(".qor-slideout__title").hide(),l.title=e(".qor-doc__preview .qor-help__document_title").text(),l.url=h,i=window.Mustache.render(t.TEMPLATE_DOC_TITLE,l),e(".qor-doc__preview .qor-help__document_title").hide(),e(".qor-slideout__header").append(i),c.attr("href",h),e(".qor-slideout__title .qor-doc__close").click(function(){r.show(),e(".qor-slideout__title").show(),e(".qor-doc__preview").remove(),e(".qor-slideout__show_title").remove(),c.attr("href",d)}),s.remove()},error:function(e,t,o){s.remove(),window.alert([t,o].join(": "))}}),!1},destroy:function(){this.unbind(),this.$element.removeData(o)}},t.TEMPLATE_LOADING='<div style="text-align: center; margin-top: 30px;"><div class="mdl-spinner mdl-js-spinner is-active qor-layout__bottomsheet-spinner"></div></div>',t.TEMPLATE_PREVIEW_CLOSE="",t.TEMPLATE_DOC_TITLE='<h3 class="qor-slideout__title qor-slideout__show_title"><a href="javascript://" class="qor-doc__close"><i class="material-icons">keyboard_backspace</i></a><span>[[title]]</span></h3>',t.plugin=function(i){return this.each(function(){var n,r=e(this),s=r.data(o);if(!s){if(/destroy/.test(i))return;r.data(o,s=new t(this,i))}"string"==typeof i&&e.isFunction(n=s[i])&&n.apply(s)})},e(function(){var o='[data-toggle="qor.help"]';e(document).on(n,function(i){t.plugin.call(e(o,i.target),"destroy")}).on(i,function(i){t.plugin.call(e(o,i.target))}).triggerHandler(i)}),t});