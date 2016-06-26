package dist

var DeployBinData string = `<html>

<head>
    <title>GoHosts v0.1</title>
    <!-- res/plus.tis -->
 <script type="text/tiscript">

const Plus = function() {
                        
  function parentModel(el) {
    for(var p = el.parent; p; p = p.parent) {
      //p.style.display;
      if(var dr = p.$model) return dr;
    }
    return null;
  }

  // setup model namespace binding, 'this' is the element
  function Model() 
  {
    var me = this;
    var mo = this.@["model"];
    this.$model = mo? eval(mo) : self.ns; 
    if(!this.$model) throw "Model {" + mo + "} not found!";
  }

  const reIsExpression=/ *[-()+*&:?|=!^"'\[\]]+ */;
  const reVarName = /[a-zA-Z_][a-zA-Z0-9_]*/;

  // 'changed()' gets called when coll[key] value has changed (updated or removed)
  function setupModelChangeHandler(coll,key,changed) 
  {
    // data change event handler 
    function on_model_change(changedef) { 
      // changedef here is:
      //[0] - symbol, one of #add,#update,#delete,#add-range,#delete-range or #update-range
      //[1] - object or vector, its element(s) was changed;
      //[2] - symbol or string, property name - for objects
      //      or start of changed range (index) in arrays  
      //[3] - end of changed range (index, not inclusive) in arrays  
      switch(changedef[0]) {
        case #update: if( key == changedef[2] ) changed(coll[key]); break; // object-property
        case #delete: if( key == changedef[2] ) changed(undefined); break; // object-property
        case #update-range: if( key >= changedef[2] &&  key < changedef[3] ) changed(coll[key]); break; // vector-index
        case #delete-range: if( key >= changedef[2] &&  key < changedef[3] ) changed(undefined); break; // vector-index
      }
    }  
    // assign it as an observer    
    Object.addObserver.call(coll,on_model_change); // subscribe to collection object changes
  }
  
  // 'added()' gets called for each new element added to coll
  function setupModelExpansionHandler(coll, added) 
  {
    // setup data object change event handler 
    function on_model_change(changedef) { 
      switch(changedef[0]) {
        case #add:        { added(coll,changedef[2]); } break; 
        case #add-range:  { var start = changedef[2]; var end = changedef[3]; for(var i = start; i < end; ++i) added(coll,i); } break; 
      }
    }      
    Object.addObserver.call(coll,on_model_change); // subscribe to collection object changes
  }
  
  // 'changed()' is called when coll.length has changed
  function setupModelLengthChangeHandler(coll, changed) 
  {
    // setup data object change event handler 
    function on_model_change(changedef) { 
      switch(changedef[0]) {
        case #add:    case #add-range:
        case #delete: case #delete-range:  changed(coll.length);  break; 
      }
    }      
    Object.addObserver.call(coll,on_model_change); // subscribe to collection object changes
  }
  
  // 'changed()' is called when coll changes
  function setupCollectionChangeHandler(coll, changed) 
  {
    // setup data object change event handler 
    function on_model_change(changedef) { changed(); }
    Object.addObserver.call(coll,on_model_change); // subscribe to collection object changes
  }
  
  
  // compile expression into function
  function compileExpr(model, expr) { 
    var func = new Function("return " + expr); func.namespace = model; return func; 
  }
  
  function setupTerminalBinding(model,thing,path,isInput,updater) 
  {
    //stdout.printf("STB %s %V\n", path, path.match( reIsExpression ) );
    if( typeof model != #object  && typeof model != #namespace)
      return;      

    if(reIsExpression.test(path)) { // looks like an expression, setup it as data observer
      var parts = path.split(reIsExpression);
      var expr = compileExpr(model,path);
      for( var part in parts) {
        if( part.match(reVarName) ) {
          var (coll,key) = Object.referenceOf(model, part);
          if(!key) continue;
          //stdout.printf("setupTerminalBinding, expr part key %s in coll %V\n", key, coll);    
          if(key == #length) setupModelLengthChangeHandler( coll, function(length) { updater(thing, expr()); } );
          else setupModelChangeHandler( coll, key, function(val) { updater(thing, expr()); } ); 
        }
      }
      thing.post(::updater(thing,expr()));
      return;
    }
     
    assert model: "Model is null while binding " + path;
    
    //stdout.printf("STB %V %V\n", model, path );
    
    var (coll,key) = Object.referenceOf(model, path);
    if( !key ) return;
    
    //stdout.printf("setupTerminalBinding: coll=%s key=%v\n", typeof coll, key);    
    
    if( key == #length ) { // special treatment for length computale property
      setupModelLengthChangeHandler( coll, function(length) { updater(thing, length); } );
      updater(thing,coll.length); // intial value
    } 
    else { // subscribe to model change notifications:
      if( isInput ) { // setup DOM change event handler 
        function on_ui_change() { coll[key] = thing.value; return false; } 
        thing.subscribe("change.plus",on_ui_change); // subscribe to the element value change event
      }
      setupModelChangeHandler( coll, key, function(val) { updater(thing,val); } );
      //WRONG: updater(thing,coll[key]);
      //we need to post this thing as it may have other aspects and behaviors
      thing.post(::updater(thing,coll[key])); // assign intial value to it;
    }
  }
  
  function valueUpdater(thing,v) { 
    //if(!thing.state.focus && v !== undefined) thing.value = v; -- v !== undefined condition breaks 2-basic-repeatable.htm case
    // when the element is in focus we are not updating it from the model - user is editing it!
    if(!thing.state.focus) thing.value = v; 
  }
  
  // setup terminal binding on 'this'
  function Terminal() {
    var path = this.@["name"];
    var thing = this; // the DOM thing
    var model = /*this.$model ||*/ parentModel(thing);
    setupTerminalBinding(model,thing,path,thing.tag != #output, valueUpdater ); 
  }
  
  const CLASS_RE = /(.*)\{\{(([-_a-z0-9]+) *\:)? *(.+)\}\}(.*)/i;
  //const CLASS_RE = /(.*)\{\{([-_a-z0-9]+)(\:)+(.+)\}\}(.*)/i;  
  
  function ClassTerminal() {
    var model = this.$model || parentModel(this);
    var thing = this; // the DOM thing
    var classattr = this.@["class"];
    assert model : "model shall exist" ;
    assert classattr : "it should not be empty";
    
    var parts = classattr.match(CLASS_RE);
    
    if( parts[1] /*&& parts[5]*/ )
      this.@["class"] = parts[1] + " " + (parts[5] || "");
    else
      this.@["class"] = undefined;
    
    var expr = parts[4];   // expression
    var cname = parts[3];  // "cls1" if class="{{cls1:expr}}", A case, expr - boolean 
                           // "" if class="{{expr}}"         , B case, expr - string  
    var sname = null;      // class name set by B case expr  
    stdout.printf("ClassTerminal %V\n", parts); 
    
    function classUpdater(thing,v) {
      thing.post(function() {
        if( cname ) {           // case A 
          stdout.println("ClassTerminal case A", cname, v || false); 
          thing.@.toggleClass(cname, v || false);  
        }
        else if( sname != v )  // case B
          thing.@.removeClass(sname); thing.@.addClass(sname = v);  
      }, true);
      //stdout.println("classUpdater", thing, v);            
    }
    setupTerminalBinding(model,thing,expr,false,classUpdater); 
  }
  
  function AttrTerminal(attrname) {
    var model = this.$model || parentModel(this);
    var thing = this; // the DOM thing
    var attrval = this.@[attrname];
    assert model : "model shall exist" ;
    assert attrval : "it should not be empty";

    var prefix = attrval ~/ "{{";
    var expr  = attrval ~% "{{"; // expression part
    var suffix = expr %~ "}}"; // class name
    expr = expr /~ "}}";  // everything in between {{...}}
    //stdout.println("A:", prefix, expr, suffix);
    
    function attrUpdater(thing,v) {
      thing.post(function() {
          thing.@[attrname] = prefix + v + suffix;  
      }, true);
    }
    setupTerminalBinding(model,thing,expr,false,attrUpdater); 
  }
  
  function boundAttributesTerminal() {
    var model = this.$model || parentModel(this);
    var thing = this; // the DOM thing
    assert model : "model shall exist" ;

    function bind(expr, attr) { 
      assert expr : "it should not be empty";
      function attrUpdater(thing,v) {
        thing.post(function() {
          thing.@[attr] = v;  
        }, true);
      }
      setupTerminalBinding(model,thing,expr,false,attrUpdater); 
    }
    
    for( var attr in this.attributes ) // scan all attributes for "@..." ones:
    {
      if( attr !like "@*" ) continue;
      var expr = this.attributes[attr];
      attr = attr.substr(1); // attribute name without '@'
      bind(expr,attr);
    }
  }
  
  function crackEachExpression(text) // parse iterator expression: each="[index,]item in collection[|extra]"
  {
    const RE = /(([a-z0-9]+)?\,?([a-z0-9]+)?) in ([^|]+)(\|(.+))?/i;  
    var parts = text.match(RE);
    if( !parts ) 
      throw "Unrecognized @each format:" + text;
   
    var nindex = parts[2]; // index var name
    var nitem = parts[3];  // 'item' var name 
    if( !nitem ) 
       (nitem,nindex) = (nindex,nitem);
    var ncoll = parts[4];
    nitem = symbol(nitem);
    if(nindex)
      nindex = symbol(nindex);
    var filter = parts.length > 6? parts[6]:null;
    
    return (ncoll, nitem, nindex, filter);
  }
  
  function eachRecord(rec,filter) { return true; }

  function functionFilter(rec,idx,filter) {
    if( filter(rec, idx) ) 
      return true;
    else
      return false;
  }

  function objectFilter(rec,idx,filter) {
    for(var (k,v) in filter) {
      var rv = rec[k];
           if( typeof v == #string ) { if( rv.indexOf(v) < 0 ) return false; }
      else if( typeof v == #function ) { if( !v(rv) ) return false;   }
      else if( rv != v ) return false;
    }
    return true;
  }
  function textFilter(rec,idx,filter) {
    for(var (k,v) in rec) {
      if( typeof v == #string && v.indexOf(filter) < 0 ) return false;
    }
    return true;
  }
  
  function Repeater() {
      var (ncoll, nitem, nindex, filter) = crackEachExpression(this.@["each"]);
      var root = parentModel(this); 
      assert root;
      var isSelect = this.tag == "select";
      var that = this; 
      var thing = isSelect? that.options: that; 
      var template = thing.first; 
      thing.clear();
      
      var renderList; // forward declaration

      var ff = eachRecord; // filter function
      var fs = null;       // filter value

      function rqRenderList() { thing.post(renderList,true); } 
     
      function appendRepeatableItem(cont, coll, idx, val, domItemIdx)
      {
        var domitem = cont[domItemIdx];
        if( domitem && domitem.$model) {
          var elval = domitem.$model[nitem];
          if( elval === val)
            return; // record already bound with the DOM  
          domitem.remove(); 
          Object.removeObserver.call(elval, rqRenderList);
        }
        //stdout.println("appendRepeatableItem", domItemIdx);
        domitem = template.clone();
        var repeatable_root = {};
        repeatable_root[nitem] = val; 
        if(nindex) repeatable_root[nindex] = idx; 
        repeatable_root.prototype = root; // repeatable root is derived from parent root
        domitem.$model = repeatable_root; // repeated element is holding local $model now for its descendannts.
        
        setupModelChangeHandler( coll, idx, 
          function(val) { if(val) repeatable_root[nitem] = val; // this will propagate changes to record content
                          else domitem.remove(); } );
        
        cont.insert(domitem,domItemIdx);
        
        if( ff !== eachRecord && typeof val == #object ) // if something in the record has changed we need to rerender the list
          Object.addObserver.call(val, rqRenderList);
      }
    
      var (rcoll,rkey) = Object.referenceOf(root, ncoll);
      var coll = rcoll[rkey];
                 //eval(ncoll,root);
      assert(coll);
      
      this.$model = coll;
     
      renderList = function () {
        //debug stacktrace;
        var seqNo = 0;
        for(var (k,v) in coll)
          if( ff(v,k,fs) )
            appendRepeatableItem(thing,coll,k,v,seqNo++);
     
        while( thing.length > seqNo ) {
          var domitem = thing.last;
          var elval = domitem.$model[nitem]; 
          domitem.remove(); 
          Object.removeObserver.call(elval, rqRenderList);
        }
        if(isSelect) // after rendering <select> items we need to reset value to match new options
          that.value = that.value;
      }
      
      if( filter ) {
        function updater(thing,val) { 
          fs = val;
          if( fs === null || fs === undefined) ff = eachRecord;
          else if( typeof fs == #function ) ff = functionFilter;
          else if( typeof fs == #object ) ff = objectFilter;
          else if( typeof fs == #string ) ff = textFilter;
          else if( !fs ) throw "Unknown filter " + fs.toString();
          renderList();
        }
        setupTerminalBinding(root,thing,filter,false,updater); 
      }
      else 
        renderList();
        
      // and setup observer for future modifications:
      setupCollectionChangeHandler(coll, rqRenderList ); 
      setupModelChangeHandler(rcoll, rkey, function() { 
         this.$model = coll = rcoll[rkey];  
         rqRenderList();
      } );
  }
  
  function handleEvent(target,name, evt) {
    var model = target.$model || parentModel(target);
    var action = target.attributes[name];
    return eval.call(target,action,model);
  }
  function handleKeyEvent(target, name, evt) {
    var code;   
    switch(name) {
      case "escape": code = Event.VK_ESCAPE; break;
      case "enter":  code = Event.VK_RETURN; break;
      default: return false;
    }
    if( evt.keyCode != code ) 
      return false;
    var model = target.$model || parentModel(target);
    var action = target.attributes[name];
    eval.call(target,action,model);
    return true;
  }
    
  function Click()    { this.subscribe("click", function(evt) { return handleEvent(this,"click",evt) }) }
  function DblClick() { this.subscribe("dblclick", function(evt) { return handleEvent(this,"dblclick",evt) }) }
  
  function Enter()    { this.subscribe("~keydown", function(evt) { return handleKeyEvent(this,"enter", evt) }) }
  function Escape()   { this.subscribe("~keydown", function(evt) { return handleKeyEvent(this,"escape", evt) }) }
  
  function FocusIn()  { this.subscribe("focusin", function(evt) { return handleEvent(this,"focusin",evt); } ) }
  function FocusOut() { this.subscribe("focusout", function(evt) { return handleEvent(this,"focusout",evt); } ) }
  
  function Change()   { this.subscribe("change", function(evt) { return handleEvent(this,"change",evt); } ) }
   
  return {
    Model:Model,
    Terminal:Terminal,
    ClassTerminal:ClassTerminal,
    AttrTerminal: AttrTerminal,
    valAttrTerminal: function() { AttrTerminal.call(this,"value"); },
    hrefAttrTerminal: function() { AttrTerminal.call(this,"href"); },
    srcAttrTerminal: function() { AttrTerminal.call(this,"src"); },
    boundAttributesTerminal: boundAttributesTerminal,
    Repeater:Repeater,
    Click:Click,
    DblClick:DblClick,
    Change:Change,
    Enter:Enter,
    Escape:Escape,
    FocusIn: FocusIn,
    FocusOut: FocusOut,
 };
 
}();


// the @observing decorator
// it gets applied to the function willing to be invoked when some data
// (defined by the path) gets changed
function @observing(func, paths..) {

  assert typeof this == #namespace;
  
  var observeChanges = this[#$observeChanges];
  if (!observeChanges)
  {
    observeChanges = function(obj, onchange, path = "") 
    {
      var typ = typeof obj;
      
      //stdout.println("observing:", path);
      if( obj && (typ == #object || typ == #namespace)) {
        var spath = path? path + "." : "";
        for(var (k,v) in obj)
          observeChanges(v, onchange, spath + k);
      }
      else if( typ == #array ) {
        var spath = path + "[]";
        for(var (k,v) in obj) 
          observeChanges(v, onchange, spath);
      }
      else 
        return;

      function subs(changedef) {
        var spath = path? path + "." : "";
        switch(changedef[0]) {
            case #add:    
            case #update: spath = spath + changedef[2]; observeChanges(obj[changedef[2]], onchange, spath); break;
            case #delete: spath = spath + changedef[2]; break; 
            case #add-range: 
            case #update-range:
            {
              var start = changedef[2], end = changedef[3];
              spath = path + "[]";
              for(var i = start; i < end; ++i)
                observeChanges(obj[i], onchange, spath);
            }
            case #delete-range: spath = path + "[]"; break;
          }
        onchange(obj,changedef,spath);
      }
          
      Object.addObserver.call(obj,subs);
    };
    
    function notifier(func,path) { return :: func(path); }
    
    observeChanges.list = [];
    observeChanges.purolator = function(obj,changedef,path) {
      for( var cb in observeChanges.list ) 
        if( path like cb.path )
           self.post( notifier(cb.func,path), true);
    }
    this[#$observeChanges] = observeChanges;
    // setup observer on the namespace
    observeChanges(this,observeChanges.purolator);
  }
  
  var onlyChanges = false;
  // subscribe
  for( var path in paths )
    if( path == #changes )
      onlyChanges = true;
    else
      observeChanges.list.push({path:path,func:func});
  // invoke it now (if not only changes requested)
  if(!onlyChanges) 
    self.post(func, true);
    
  return func;
}


/* NOTE: this method is implemented natively now
// that is an equivalent of jQuery $.extend(deep) method:
function Object.extend(objects..) 
{
  const extend1 = function(reciever,source) {
    var rv;
    for( var (k,v) in source ) {
      if( typeof v == #object && typeof (rv = reciever[k]) == #object )
        extend1(rv,v);
      else
        reciever[k] = v;
    }
  }
  for( var obj in objects )
    extend1(this,obj);
}*/


</script>

    <!-- res/colorizer.tis -->
 <script type="text/tiscript">

function colorize() 
{
  const apply = Selection.applyMark; // shortcut
	const isEditor = this.tag == "plaintext";
  
  // forward declarations:
  var doStyle;
  var doScript;

  // markup colorizer  
  function doMarkup(tz, val) 
  {
		var bnTagStart = null;
		var tagScript = false;
		var tagScriptType = false;
		var tagStyle = false;
		var textElement;
      
    while(var tt = tz.token()) {
      if( isEditor && tz.element != textElement )       
      {
         textElement = tz.element;
         textElement.attributes["type"] = "markup";
      }
      //stdout.println(tt,tz.attr,tz.value);
      switch(tt) {
        case #TAG-START: {    
            bnTagStart = tz.tokenStart; 
            const tag = tz.tag;
            tagScript = tag == "script";
            tagStyle  = tag == "style";
          } break;
        case #TAG-HEAD-END: {
            apply(bnTagStart,tz.tokenEnd,"tag"); 
            if( tagScript ) { tz.push(#source,"</" + "script>"); doScript(tz, tagScriptType, true, val); }
            else if( tagStyle ) { tz.push(#source,"</" + "style>"); doStyle(tz, true, val); }
          } break;
        case #TAG-END:      apply(tz.tokenStart,tz.tokenEnd,"tag"); break;  
        case #TAG-ATTR:     if( tagScript && tz.attr == "type") tagScriptType = tz.value; 
                            if( tz.attr == "id" ) apply(tz.tokenStart,tz.tokenEnd,"tag-id"); 
                            break;
      }
    }
  }
 
  // script colorizer
  doScript = function(tz, typ, embedded = false, val) 
  {
    const KEYWORDS = 
    {
      "type"    :true, "function":true, "var"       :true,"if"       :true,
      "else"    :true, "while"   :true, "return"    :true,"for"      :true,
      "break"   :true, "continue":true, "do"        :true,"switch"   :true,
      "case"    :true, "default" :true, "null"      :true,"super"    :true,
      "new"     :true, "try"     :true, "catch"     :true,"finally"  :true,
      "throw"   :true, "typeof"  :true, "instanceof":true,"in"       :true,
      "property":true, "const"   :true, "get"       :true,"set"      :true,
      "include" :true, "like"    :true, "class"     :true,"namespace":true,
      "this"    :true, "assert"  :true, "delete"    :true,"otherwise":true,
      "with"    :true, "__FILE__":true, "__LINE__"  :true,"__TRACE__":true,
      "debug"   :true, "await"   :true 
    };
      
    const LITERALS = { "true": true, "false": true, "null": true, "undefined": true };
    
    var firstElement;
    var lastElement;
    var commentLine = null;

    while:loop(var tt = tz.token()) {
      var el = tz.element;
      if( !firstElement ) firstElement = el;
      lastElement = el;

      if (commentLine != null &&  commentLine == tz.tokenStart[0].text.substr(0, commentLine.length)) {
        apply(tz.tokenStart,tz.tokenEnd, "comment"); 
      } else {
        commentLine = null;
        switch(tt) 
        {
          case #NUMBER:       apply(tz.tokenStart,tz.tokenEnd,"number"); break; 
          case #NUMBER-UNIT:  apply(tz.tokenStart,tz.tokenEnd,"number-unit"); break; 
          case #STRING:       apply(tz.tokenStart,tz.tokenEnd,"string"); break;
          case #NAME:         
          {
            var val = tz.value;
            if( val[0] == '#' ) {
              commentLine = tz.tokenStart[0].text;
              apply(tz.tokenStart,tz.tokenEnd, "comment"); 
            }
            else if(KEYWORDS[val]) 
              apply(tz.tokenStart,tz.tokenEnd, "keyword"); 
            else if(LITERALS[val]) 
              apply(tz.tokenStart,tz.tokenEnd, "literal"); 
            break;
          }
          case #COMMENT:      apply(tz.tokenStart,tz.tokenEnd,"comment"); break;
          case #END-OF-ISLAND:  
            // got script end tag
            tz.pop(); //pop tokenizer layer
            break loop;
        }
      }
    }
    if(isEditor && embedded) {
      for( var el = firstElement; el; el = el.next ) {
        el.attributes["type"] = "script";
        if( el == lastElement )
          break;
      }
    }
  };
  
  doStyle = function(tz, embedded = false, val) 
  {
    const KEYWORDS = 
    {
      "rgb":true, "rgba":true, "url":true, 
      "@import":true, "@media":true, "@set":true, "@const":true
    };
      
    const LITERALS = { "inherit": true };
    
    var firstElement;
    var lastElement;
    
    while:loop(var tt = tz.token()) {
      var el = tz.element;
      if( !firstElement ) firstElement = el;
      lastElement = el;
      switch(tt) 
      {
        case #NUMBER:       apply(tz.tokenStart,tz.tokenEnd,"number"); break; 
        case #NUMBER-UNIT:  apply(tz.tokenStart,tz.tokenEnd,"number-unit"); break; 
        case #STRING:       apply(tz.tokenStart,tz.tokenEnd,"string"); break;
        case #NAME:         
        {
          var val = tz.value;
          if( val[0] == '#' )
            apply(tz.tokenStart,tz.tokenEnd, "symbol"); 
          else if(KEYWORDS[val]) 
            apply(tz.tokenStart,tz.tokenEnd, "keyword"); 
          else if(LITERALS[val]) 
            apply(tz.tokenStart,tz.tokenEnd, "literal"); 
          break;
        }
        case #COMMENT:      apply(tz.tokenStart,tz.tokenEnd,"comment"); break;
        case #END-OF-ISLAND:  
          // got  script end tag
          tz.pop(); //pop tokenizer layer
          break loop;
      }
    }
    if(isEditor && embedded) {
      for( var el = firstElement; el; el = el.next ) {
        el.attributes["type"] = "style";
        if( el == lastElement )
          break;
      }
    }
  };
  
  var me = this;
  
  function doIt() { 
  
    var typ = me.attributes["type"];
    var val = me.state.value;
      
    var syntaxKind = typ like "*html" || typ like "*xml" ? #markup : #source;
    var syntax = typ like "*css"? #style : #script;
      
    var tz = new Tokenizer( me, syntaxKind );
  
    if( syntaxKind == #markup )
      doMarkup(tz, val);
    else if( syntax == #style )
      doStyle(tz, false, val);
    else 
      doScript(tz,typ, false, val);
  }
  
  doIt();
  
  // redefine value property
  this[#value] = property(v) {
    get { return this.state.value; }
    set { this.state.value = v; doIt(); }
  };
  
  this.load = function(text,sourceType) 
  {
    this.attributes["type"] = sourceType;
    if( !isEditor )
      text = text.replace(/\r\n/g,"\n"); 
    this.state.value = text; 
    doIt();
  };
  
  this.sourceType = property(v) {
    get { return this.attributes["type"]; }
    set { this.attributes["type"] = v; doIt(); }
  };
  if (isEditor)
		this.on("change", function() {
			this.timer(40ms,doIt);
		});
 

}
</script>

    <!-- res/notification.tis -->
 <script type="text/tiscript">

class Notification : Behavior 
{
	 const WAIT_TIMER = 5s; // duration to show it

	 var msg; // singleton, one message at the same time
     
	 function cancel() { // call if you want to cancel it
		 if( msg.shown ) {
      msg.off(".notification");
			msg.attributes["state"] = undefined;
			msg.move();
		 }		 
	 }
   
  function show(message, title = "Some message:") { 
   
		cancel();
		msg.$( content ).html = message;
		msg.$( header ).text = title;

    // NOTE: state animations are defined in CSS
    function closeIt() {
      if(!msg.isVisible) return;
      msg.attributes["state"] = "closed";
      msg.on("animationend.notification", function(evt) { 
        msg.off(this function); // do it once
        msg.attributes["state"] = undefined;
        msg.move();
        return true;
      });
    }

    function revealIt() {
      if(!msg.isVisible) return;
      msg.attributes["state"] = "shown";
      msg.on("animationend.notification", function(evt) { 
        if(!msg.isVisible) return;
        msg.off(this function); // do it once
        msg.timer( WAIT_TIMER, closeIt );
        return true;
      });
    }

    // getting this monitor's box
    var (screenX1,screenY1,screenX2,screenY2) = view.screenBox(#workarea, #rect );

    // set initial "off stage" state 
    msg.attributes["state"] = "initial";
    
    // measure its real sizes
    var (w,h) = msg.box(#dimension,#margin);
    var (oleft,otop,oright,obottom) = msg.box(#rect,#margin,#inner);
    //stdout.println(oleft,otop,oright,obottom);
    
    // position it at bottom / right corner
		msg.move( screenX2 - w, 
              screenY2 - h, 
              #screen, 
              #detached-topmost-window );
    // request to start animation
    msg.post( revealIt );
		
  }
     
  function attached() {
		msg = this;
	}

}

</script>

    <style>
      /* res/plus.css */

/* principal handlers */
[model] { aspect:"Plus.Model"; }
[model] [each] { aspect:"Plus.Repeater"; } /* note repeater shall come first before [model] [name] for <select each="..." name="var"> */
[model] [name] { aspect:"Plus.Terminal"; }
[model] [class*='{{'] { aspect:"Plus.ClassTerminal"; }
[model] [value*='{{'] { aspect:"Plus.valAttrTerminal"; }
[model] [href*='{{'] { aspect:"Plus.hrefAttrTerminal"; }
[model] [src*='{{'] { aspect:"Plus.srcAttrTerminal"; }

/* any attribute with the name starting from '@' is considered as bound: */ 
[model] *:has-bound-attributes { aspect:"Plus.boundAttributesTerminal"; }

/* auxiliary event handlers */
[model] [click] { aspect:"Plus.Click"; }
[model] [dblclick] { aspect:"Plus.DblClick"; }
[model] [change] { aspect:"Plus.Change"; }
[model] [enter] { aspect:"Plus.Enter"; }
[model] [escape] { aspect:"Plus.Escape"; }
[model] [focusin] { aspect:"Plus.FocusIn"; }
[model] [focusout] { aspect:"Plus.FocusOut"; }




      /* res/colorizer.css */

@set colorizer < std-plaintext 
{
  :root { aspect: colorize; }
  
  text { white-space:pre; }
  /*markup*/  
  text::mark(tag) { color: olive; } /*background-color: #f0f0fa;*/
  text::mark(tag-id) { color: red; } /*background-color: #f0f0fa;*/

  /*source*/  
  text::mark(number) { color: brown; }
  text::mark(number-unit) { color: brown; }
  text::mark(string) { color: teal; }
  text::mark(keyword) { color: blue; }
  text::mark(symbol) { color: brown; }
  text::mark(literal) { color: brown; }
  text::mark(comment) { color: green; }
  
  text[type=script] {  background-color: #FFFAF0; }
  text[type=markup] {  background-color: #FFF;  }
  text[type=style]  {  background-color: #FAFFF0; }
}

plaintext[type] {
  style-set: colorizer;
}

@set element-colorizer 
{
  :root { 
	  aspect: colorize; 
	  background-color: #fafaff;
		padding:4dip;
		border:1dip dashed #bbb;
	}
  
  /*markup*/  
  :root::mark(tag) { color: olive; } 
  :root::mark(tag-id) { color: red; }

  /*source*/  
  :root::mark(number) { color: brown; }
  :root::mark(number-unit) { color: brown; }
  :root::mark(string) { color: teal; }
  :root::mark(keyword) { color: blue; }
  :root::mark(symbol) { color: brown; }
  :root::mark(literal) { color: brown; }
  :root::mark(comment) { color: green; }
}

pre[type] {
  style-set: element-colorizer;
}


      /* res/notification.css */

popup#notification { style-set: notification-message; }

@set notification-message {

  :root {
    prototype: Notification;
    margin:20dip;
    width: 300dip;
    display:none;
    transform: translate(0,0); opacity:1.0;
    font:system;
    background-color: #fff;
    border-color: 1px solid #999;
  }
  
  logo { 
      foreground: url(path:M1664 640q53 0 90.5 37.5t37.5 90.5-37.5 90.5-90.5 37.5v384q0 52-38 90t-90 38q-417-347-812-380-58 19-91 66t-31 100.5 40 92.5q-20 33-23 65.5t6 58 33.5 55 48 50 61.5 50.5q-29 58-111.5 83t-168.5 11.5-132-55.5q-7-23-29.5-87.5t-32-94.5-23-89-15-101 3.5-98.5 22-110.5h-122q-66 0-113-47t-47-113v-192q0-66 47-113t113-47h480q435 0 896-384 52 0 90 38t38 90v384zm-128 604v-954q-394 302-768 343v270q377 42 768 341z)  no-repeat 50% 50%; 
      foreground-size:34dip;
      size:34dip;
      fill:#000;
      background-color: #EBEBEB; display: block; font-size: 34dip; width: 100dpi; float: left; padding: 18dip;
  }
  header { display:block; font-weight: bold; padding: 10dip 10dip 5dip 10dip; margin-left: 80dip;  margin-right: 10dip; font-size: 16dip; border-bottom: 1px dashed gray;}
  content { display:block; padding: 5dip 10dip 10dip 10dip; margin-left: 80dip; margin-right: 10dip; font-size: 14dip;}


  :root[state] {display:block;}
  :root[state=initial] { transform: translate(0,100%); opacity:0.0; } /* offstage */
  :root[state=shown]   { transform: translate(0,0);    opacity:1.0; transition: transform(quad-out,500ms) opacity(linear,500ms);  }
  :root[state=closed]  { transform: translate(100%,0); opacity:0.0; transition: transform(linear,500ms) opacity(linear,500ms);  }
  

}

      /* res/app.css */
  body {
    flow:horizontal; /* layout style: http://www.terrainformatica.com/htmlayout/flow.whtm */
    padding: 0px;
    margin:0px;
    font: 14dip;
  }

  #menu {
    width:25%%; 
    min-width: 240dip;
    height: *;
    background-color: #373D47;
    color:#979DA7;
    fill:#979DA7;
    padding: 5dip 0dip;
    border-top: solid 1px #C5C3C5;
    position: relative;
  }
  #menu ul{ 
      overflow-y: auto;
      padding: 0px;
      margin-bottom:30dip;
      height: *;
    }

    #menu li {
      display:block; 
      padding:0 20dip;
      line-height:31dip;
      text-decoration: none;
      list-style: none;
      overflow: hidden;
    }
    #menu li:hover {
        background-color: #2C3139;
    }
    #menu li.active {
        background-color: #2C3139;
        color:#fff;
    }
    #menu .icon {
      margin-right:5dip;
      display: inline-block;
    }
    #menu .toggle-icon {
      float: right;
      width: 16dip;
      display: inline-block;
      margin-top: 4dip;
    }
    #menu ul span {
      size:14dip;
      padding:0;
      stroke:none; 
    }
    #menu ul span.icon {
      foreground: url(path:M1596 380q28 28 48 76t20 88v1152q0 40-28 68t-68 28h-1344q-40 0-68-28t-28-68v-1600q0-40 28-68t68-28h896q40 0 88 20t76 48zm-444-244v376h376q-10-29-22-41l-313-313q-12-12-41-22zm384 1528v-1024h-416q-40 0-68-28t-28-68v-416h-768v1536h1280zm-1024-864q0-14 9-23t23-9h704q14 0 23 9t9 23v64q0 14-9 23t-23 9h-704q-14 0-23-9t-9-23v-64zm736 224q14 0 23 9t9 23v64q0 14-9 23t-23 9h-704q-14 0-23-9t-9-23v-64q0-14 9-23t23-9h704zm0 256q14 0 23 9t9 23v64q0 14-9 23t-23 9h-704q-14 0-23-9t-9-23v-64q0-14 9-23t23-9h704z)  no-repeat 50% 50%; 
      foreground-size:10dip;
    }
    #menu li:first-child span.icon {
      foreground: url(path:M1728 992v-832q0-13-9.5-22.5t-22.5-9.5h-1600q-13 0-22.5 9.5t-9.5 22.5v832q0 13 9.5 22.5t22.5 9.5h1600q13 0 22.5-9.5t9.5-22.5zm128-832v1088q0 66-47 113t-113 47h-544q0 37 16 77.5t32 71 16 43.5q0 26-19 45t-45 19h-512q-26 0-45-19t-19-45q0-14 16-44t32-70 16-78h-544q-66 0-113-47t-47-113v-1088q0-66 47-113t113-47h1600q66 0 113 47t47 113z)  no-repeat 50% 50%; 
      foreground-size:14dip;
    }
    #menu ul li.active span.icon {
      color:#fff;
      fill: #fff;
    }
    #menu li:not(:first-child) span.toggle-icon {
      foreground: url(path:M1024 896q0-104-40.5-198.5t-109.5-163.5-163.5-109.5-198.5-40.5-198.5 40.5-163.5 109.5-109.5 163.5-40.5 198.5 40.5 198.5 109.5 163.5 163.5 109.5 198.5 40.5 198.5-40.5 163.5-109.5 109.5-163.5 40.5-198.5zm768 0q0-104-40.5-198.5t-109.5-163.5-163.5-109.5-198.5-40.5h-386q119 90 188.5 224t69.5 288-69.5 288-188.5 224h386q104 0 198.5-40.5t163.5-109.5 109.5-163.5 40.5-198.5zm128 0q0 130-51 248.5t-136.5 204-204 136.5-248.5 51h-768q-130 0-248.5-51t-204-136.5-136.5-204-51-248.5 51-248.5 136.5-204 204-136.5 248.5-51h768q130 0 248.5 51t204 136.5 136.5 204 51 248.5z)  no-repeat 50% 50%; 
      foreground-size:16dip;
    }

    #menu ul span.toggle-icon.on {
      /*foreground: url(path:M-128 896q0-130 51-248.5t136.5-204 204-136.5 248.5-51h768q130 0 248.5 51t204 136.5 136.5 204 51 248.5-51 248.5-136.5 204-204 136.5-248.5 51h-768q-130 0-248.5-51t-204-136.5-136.5-204-51-248.5zm1408 512q104 0 198.5-40.5t163.5-109.5 109.5-163.5 40.5-198.5-40.5-198.5-109.5-163.5-163.5-109.5-198.5-40.5-198.5 40.5-163.5 109.5-109.5 163.5-40.5 198.5 40.5 198.5 109.5 163.5 163.5 109.5 198.5 40.5z) no-repeat 50% 50%;  */
      foreground: url(path:M1671 566q0 40-28 68l-724 724-136 136q-28 28-68 28t-68-28l-136-136-362-362q-28-28-28-68t28-68l136-136q28-28 68-28t68 28l294 295 656-657q28-28 68-28t68 28l136 136q28 28 28 68z)  no-repeat 50% 50%; 
      foreground-size:16dip;
      fill:#ABFF99;
    }

    

    /*#menu ul span.icon.system::after {
      font: 14dip MyAwesome;
      content:"&fa-desktop;";
    }*/


    #toolbar-left {
        position: absolute;
        bottom: 0px;
        left:0px;
        width: 100dpi;
        padding: 10dip 20dip;
    }
    #toolbar-right {
        position: absolute;
        bottom: 0px;
        right:0px;
        width: 100dpi;
        padding: 10dip 20dip;
    }
    #toolbar-right span {
        margin-left: 10dip;
    }
    #toolbar-left span,#toolbar-right span {
        font-size: 20dip !important;
    }
    #toolbar-left span:hover,#toolbar-right span:hover {
        color:#fff;
        fill:#fff;
    }

    #toolbar-left span {
      foreground: url(path:M1600 736v192q0 40-28 68t-68 28h-416v416q0 40-28 68t-68 28h-192q-40 0-68-28t-28-68v-416h-416q-40 0-68-28t-28-68v-192q0-40 28-68t68-28h416v-416q0-40 28-68t68-28h192q40 0 68 28t28 68v416h416q40 0 68 28t28 68z)  no-repeat 50% 50%; 
      foreground-size:14dip;
      size:16dip;
      padding:0;
      stroke:none; 
    }

    #toolbar-right span {
      foreground: url(path:M704 736v576q0 14-9 23t-23 9h-64q-14 0-23-9t-9-23v-576q0-14 9-23t23-9h64q14 0 23 9t9 23zm256 0v576q0 14-9 23t-23 9h-64q-14 0-23-9t-9-23v-576q0-14 9-23t23-9h64q14 0 23 9t9 23zm256 0v576q0 14-9 23t-23 9h-64q-14 0-23-9t-9-23v-576q0-14 9-23t23-9h64q14 0 23 9t9 23zm128 724v-948h-896v948q0 22 7 40.5t14.5 27 10.5 8.5h832q3 0 10.5-8.5t14.5-27 7-40.5zm-672-1076h448l-48-117q-7-9-17-11h-317q-10 2-17 11zm928 32v64q0 14-9 23t-23 9h-96v948q0 83-47 143.5t-113 60.5h-832q-66 0-113-58.5t-47-141.5v-952h-96q-14 0-23-9t-9-23v-64q0-14 9-23t23-9h309l70-167q15-37 54-63t79-26h320q40 0 79 26t54 63l70 167h309q14 0 23 9t9 23z)  no-repeat 50% 50%; 
      foreground-size:14dip;
      size:16dip;
      padding:0;
      stroke:none; 
    }

  #editor {
    width: 75%%;
    border: solid 1px #C5C3C5;
    height: *;
  }
  .hide {
    display: none !important;
  }

  #overlay {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    background: #000;
    opacity: 0.618;
    z-index: 90;
}

#edit-form {
    position: fixed;
    top: 20%;
    left: 50%;
    background: #fff;
    z-index: 100;
    width: 420dip;
    margin-left: -200dip;
}

#edit-form h2 {
    background: #f5f5f5;
    font-size: 14dip;
    font-weight: normal;
    padding: 0 10dip;
    margin: 0;
    line-height: 40dip;
}

#edit-form .form-group label {
    display: inline-block;
    width: 80dip;
}

#edit-form .form-group input {
    width: 240dip;
    padding: 6dip 10dip;
    outline: none;
}

#edit-form .form-group {
    padding: 20dip;
    line-height: 34dip;
}

#edit-form .form-footer {
        background: #f5f5f5;
padding: 16px 20px;
text-align: right;
}

#edit-form button{
    color: #333;
    background: #fff;
    border: solid 1px #ccc;
    padding: 2dip 16dip;
    margin-left: 1em;
    cursor: pointer;
    font-size: 14dip;
    color: #000;
}

#edit-form  button.btn-default {
    color: #333;
    background: #fff;
    /*border: #ccc;*/
}

#edit-form  button.btn-submit {
    background: #09f;
    color: #fff;
}

#msg {
    color: red;
    padding: 8dip 0 8dip 85dip;
}

plaintext > text {
    margin-left:3em;
    hit-margin: 0 0 0 3em;
    white-space: pre;
    display: list-item;
    list-style-type: index;  
}

#searchbox {
    position: absolute;
    right: 0px;
    top: 0px;
    z-index: 99;
}

#searchbox input {
    padding-right:64dip;
    width: 100dip;
}

#searchbox button {
    margin-left: -6px;
}

#searchtip {
    position: absolute;
    right: 30dip;
    top: 2px;
    color: #999;
    z-index: 100;
}

#searchtip>span.noitems {
    color: red;
}

.with-context-menu
{
  context-menu: selector(menu#context);
}

#edit-form h2::after {
  content: "Add New Host";
}

#edit-form h2.edit::after {
  content: "Edit Host";
}

    </style>
</head>

<body model="App">

    <div id="menu">
        <ul each="index,host in hosts">
            <li click="showHost(host)" dblclick="switchHost(host)" class="with-context-menu {{active:host.active}}" @data-name="host.name">
            <span class="icon" title="drag to change order"> </span>
            <output(host.title)/>
            <span class="toggle-icon {{on:host.icon == #host && host.on}}"> </span>
            </li>
        </ul>

        <div id="toolbar-left">
            <span class="icon" click="editHost()"> </span>
        </div>
        <div id="toolbar-right">
            <span class="icon" click="deleteHost()" > </span>
        </div>
 
  
    </div>

    <div id="editor">
        <plaintext type="text/hosts" spaces-per-tab=2 change="changeHost()" ></plaintext>
        <div id="searchbox" class="hide">
            <input type="text" name="">
            <button click="hideSearchBox()">X</button>
        </div>
        <div id="searchtip" class="{{hide:searchItemsCount == -1}}">
        <span class="{{noitems:searchItemsCount == 0}}">
        <output(searchItemsCount)/> items
        </span>
        </div>
    </div>


    <popup#notification>
      <logo></logo>
      <header>Question:</header>
      <content></content>
    </popup>

    <div class="hide">
    <menu.context id="context">
        <li id="context_edit_host">Change Host Title</li>
        <li id="context_delete_host">Delete Host</li>
    </menu>
    </div>

    <div id="overlay" class="hide"></div>
    <div id="edit-form" class="hide">
        <h2></h2>
        <form class="form-horizontal">
            
            <div class="form-group">
                <label>Host Title: </label>
                <input|text(editHostTitle) />
                <div id="msg" class="hide">title cannot be empty.</div>
            </div>
            <div class="form-footer">
                <button class="btn" click="cancelEdit()">Cancel</button>
                <button class="btn btn-submit" click="doEditHost()">Save</button>
            </div>
        </form>
    </div>


    <!-- res/ddm.tis -->
 <script type="text/tiscript">

// drag-n-drop manager, setup draggable environment
// Example of the call of this function (using "call-object" notation):
// DragDrop
//    {
//      what      : "ul#tools > li, ul.zone > li",
//      where     : "ul.zone",
//      notBefore : "ul.zone > caption",
//      acceptDrag: dragType,
//      easeDrop  : Animation.Ease.OutQuad
//    };

function DragDrop(def) 
{
  //| 
  //| def here is an object that has following fields:
  //|
  //| def.what = selector [string], defines group of draggable elements.
  //| def.where = selector [string], defines group of target elements where dragables can be dropped.
  //| def.notBefore = selector [string], defines positions where drop is not allowed. 
  //| def.acceptDrop = function(draggable, target), function to be called before the drop, if it is defined and returns true operation ends successfully.
  //| def.acceptDrag = function(draggable), function to be called before the drag starts, if it is defined and returns either #copying or #moving operation starts successfully.
  //| def.arrivedTo = function(draggable, target), function to be called when draggable enters the target.
  //| def.dropped = function(draggable, from), function to be called when draggable dropped on the target. 
  //|               target is new draggable.parent and 'from' is a previous parent of the draggable.
  //| def.container = parent-selector [string], selector of the nearest parent of def.what elements where DD operation is allowed.
  //| def.easeDrop = function(t, b, c, d) - ease function used for the drop animation, one of Animation.Ease.*** functions.
  //| def.setupPlaceholder = function(placeholderElement) - do something special with created placeholder.
  //| def.animationDuration = milliseconds, duration of "docking" animation
  //| def.before = function(), called before entering DD loop
  //| def.after = function(), called after finshing DD loop
  //| def.autoScroll = true | false , if autoScroll of container is required

  const X_THRESHOLD = 4;
  const Y_THRESHOLD = 4;
  const STEP_DELAY  = 8;
  const PING_THRESHOLD = 400ms;
  const ANI_DURATION = def.animationDuration || 200ms;

  var dd_x, dd_y;
  var dd_op = #moving;
  var dd_source = null;     // the draggable
  var dd_target = null;     // current target, mouse is over it.
  var dd_targets = def.where instanceof Element ? [def.where] : self.selectAll(def.where);
  var dd_placeholder_src = null;
  var dd_placeholder_dst = null;
  var dd_container = null;  // DD happens inside this only
  var dd_width, dd_height;  // dims of the draggable
  var is_animating;
  var requested_cancel = false;
  var dd_dragging = null;
  var dd_autoScroll = def.autoScroll !== undefined ? def.autoScroll : true;
  var dd_movable = def.movable;
  var dd_movable_el = null;

 
  // forward declaration of functions:
  var doDrop; 
  var doCancelDrop; 
  var setupDstPlaceholderAt; 
  var setupSrcPlaceholderAt; 
  //var onMouseHandler;
    
  // do cleanup
  function ddFinalize()
  {
    // clean all this up
    for(var tel in dd_targets)
      tel.state.droptarget = false; 
    if(dd_target) dd_target.state.dragover = false; 
    if(dd_placeholder_dst) dd_placeholder_dst.remove();
    if(dd_placeholder_src) dd_placeholder_src.remove();
    if(dd_source) dd_source.state[ dd_op ] = false;
    
    // Be polite with the GC:
    dd_target = dd_placeholder_src = dd_placeholder_dst = dd_source = null;
  }
  
  // init-loop-commit:
  function doDD(el, vx, vy) // DD loop
  {
   // 1) ask initiator about our draggable:
    if( def.acceptDrag ) 
    {
      dd_op = def.acceptDrag( el );
      if( dd_op != #copying && dd_op != #moving )
        return false; // not this time, sigh.
    }
  // 1-bis) setup container, if any:  
    if( def.container )
    {
      dd_container = el.parent.selectParent(def.container);
      assert dd_container;
    }
  
  // 2) find and mark all allowed targets:
    dd_targets = def.where instanceof Element ? [def.where] : self.selectAll(def.where);
    
  // sort all dd_targets by depth, so child options can be found before the whole <select>
    dd_targets.sort(:e1,e2{
      function depth(e)
      {
        var depth = 0;
        do { e = e.parent; depth++; } while(e.parent);
        return depth;
      }
      var d1 = depth(e1);
      var d2 = depth(e2);
      if( d1 < d2 ) return 1;
      if( d1 === d2 ) return 0;
      return -1;

    });

    assert dd_targets.length > 0;
    for(var tel in dd_targets) 
      tel.state.droptarget = true; // to give CSS a chance to highlight them somehow using :drop-target
    
    dd_source = el;
    (dd_width, dd_height) = el.box(#dimension);

  // 3) create placeholder of the draggable, it will hold its place: 
    if( dd_op == #moving )
    {
      dd_placeholder_src = dd_source.tag == "tr" ? dd_source.clone() : new Element(dd_source.tag,"");
    } else { // copying 
      dd_placeholder_src = dd_source.clone();
    }
    dd_placeholder_src.@.addClass("placeholder","src");
  // 3.a) append placeholder to the end of dd_source.parent:
    dd_source.parent.insert(dd_placeholder_src); 
    
    dd_placeholder_src.style.set { width:px(dd_width), height:px(dd_height) };
    
  // 3.b) exchange positions of dd_source and dd_placeholder_src so dd_source 
  //      that we move will always be at the end so it will not conflict with findByPos
    dd_source.swap(dd_placeholder_src); 
    
  // 3.c) call def.setupPlaceholder for dd_placeholder_src so caller can do something special with it. 
    if (def.setupPlaceholder)
      def.setupPlaceholder(dd_placeholder_src);
    
  // 4) mark the draggable and take it off: 
  
    dd_source.state[ dd_op ] = true;
    
    assert dd_width && dd_height;
    
    dd_source.move( vx - dd_x, vy - dd_y, dd_width, dd_height/*, #view, #detached-window*/);
    
  // 4a) call user's preparation code      
    if(def.before) 
      def.before();
    
  // 5) commit screen updates:
    view.update();
   
  // 6) DD events until mouse up is received
    requested_cancel = false;
    dd_dragging = el;
    el.capture(#strict);
    
    if(!view.doEvent(#untilMouseUp))
      requested_cancel = true;

    el.capture(false);
    dd_dragging = null;
     
  // 7) Loop finished, do either drop or cancel it:   
  
    if( !requested_cancel && dd_target && dd_source) 
      doDrop();
    else if(dd_source)
      doCancelDrop();

  // 7) run user's finalizer   
    if(def.after) 
      el.post(def.after);
   
    return true;
  }
  
  function findRowRange(vy) {
    var nrows = dd_target.rows;
    var top = dd_target.box(#top, #inner, #view);
    vy -= top;
    var firstIdx = 0, lastIdx = 0; 
    for( var r = 0; r < nrows; ++r )
    {
      var els = dd_target.row(r);
      firstIdx = els.first.index;
      lastIdx = els.last.index;
      var (ry,rh) = dd_target.rowY(r);
      if( vy < ry + rh)
        break; 
    }
    return (firstIdx, lastIdx);
  }

  function findColRange(vx) {
    var ncols = dd_target.columns;
    var left = dd_target.box(#left, #inner, #view);
    vx -= left;
    var firstIdx = 0, lastIdx = 0; 
    for( var c = 0; c < ncols; ++c )
    {
      var els = dd_target.column(c);
      firstIdx = els.first.index;
      lastIdx = els.last.index;
      var (cx,cw) = dd_target.columnX(c);
      if( vx < cx + cw)
        break; 
    }
    return (firstIdx, lastIdx);
  }

  
  function findPosHorz(vx,vy,multiRow = false) 
  {
    var notb = def.notBefore;
    var firstIdx = 0; 
    var lastIdx = dd_target.length - 1; /*non inclusive*/    
    
    if(multiRow) 
      (firstIdx,lastIdx) = findRowRange(vy);
    
    if( dd_target == dd_source.parent ) // 
      --lastIdx; // exclude current source element

    if( firstIdx > lastIdx )
      return firstIdx;
      
    var i;
    for( i = firstIdx; i <= lastIdx; ++i ) 
    {
      var tc = dd_target[i];
      var (x1,y1,x2,y2) = tc.box(#rect, #margin, #view);
      if(  vx < ((x1 + x2) / 2) ) {
        if(!notb || !tc.match(notb)) 
          return i;
      }
    }
    return i;
  }

  function findPosVert(vx,vy,multiCol = false) 
  {
    var notb = def.notBefore;
    var firstIdx = 0; 
    var lastIdx = dd_target.length - 1; /*non inclusive*/    
    
    if(multiCol) 
      (firstIdx,lastIdx) = findColRange(vx);
    
    if( dd_target == dd_source.parent ) // 
      --lastIdx; // exclude current source element

    if( firstIdx > lastIdx )
      return firstIdx;    

    var i;
    for( i = firstIdx; i <= lastIdx; ++i) 
    {
      var tc = dd_target[i];
      var (x1,y1,x2,y2) = tc.box(#rect, #margin, #view);
      if(  vy < ((y1 + y2) / 2) ) {
        if(!notb || !tc.match(notb)) 
          return i;
      }
    }
    return i;
  }
  
  function validPosition(index) {
    if(!def.notBefore) return index;
    if(index >= dd_target.length ) return index;
    if(!dd_target[index].match(def.notBefore)) return index;
    return null;
  }
  
  function findPosWrap(vx,vy,vert) 
  {
    var notb = def.notBefore;
    
    var (tvx, tvy) = dd_target.box(#position,#inner,#view);
    var tc = dd_target.find(vx - tvx, vy - tvy);
    
    while(tc && tc.parent !== dd_target )
      tc = tc.parent;

    //var tc = view.root.find(vx, vy);
    
    if( tc && tc.parent === dd_target ) {
      if( tc.index <= dd_placeholder_src.index || dd_target !== dd_source.parent )
        return validPosition(tc.index);
      else 
        return validPosition(tc.index + 1);
    }
    return dd_target.length;
  }
  
  
  function doMove(vx, vy)
  {
   //   stdout.$n({vx} {vy});
    if( !dd_source ) return;
    
    var x = vx - dd_x;
    var y = vy - dd_y;
    // move the draggable:
    if( dd_container )
    {
      var (x1,y1,x2,y2) = dd_container.box(#rect, #inner, #view);
      var (mx1,my1,mx2,my2) = dd_source.box(#rect, #margin, #inner); // actual margin sizes of the draggable
      var (w,h) = dd_source.box(#dimension, #inner); // actual dimensions of the draggable
      // inflate container rect:
      x1 += mx1; x2 -= mx2; 
      y1 += my1; y2 -= my2; 
      // apply positioning constraints we've got:
      if( x < x1 ) x = x1; else if( x + w > x2 ) x = x2 - w + 1;
      if( y < y1 ) y = y1; else if( y + h > y2 ) y = y2 - h + 1;
      vy = y; vx = x;
    }
    
    dd_source.move(x, y, dd_width, dd_height /*, #view, #detached-window*/);
  
    var found = null;
    for( var tel in dd_targets )
    {
      var (x1,y1,x2,y2) = tel.box(#rect, #inner, #view);
      if( vx >= x1 && vy >= y1 && vx <= x2 && vy <= y2 ) { found = tel;  break; }
    }
    //stdout.$n({found.tag});
    if( dd_target !== found )
    {
      if(dd_target) // we have left it
      { 
        dd_target.state.dragover = false; // CSS: :drag-over 
        if( dd_placeholder_dst ) { 
          dd_placeholder_dst.detach(); dd_placeholder_dst = null; 
        }
      }
      dd_target = found;
      if(dd_target) dd_target.state.dragover = true;
    }
   
    if(!dd_target) 
      return;
    
    // ok, we are on dd_target, find insert position on it
    var flow = dd_target.style#flow;
    var horz = false;
    var pos = 0;

    switch( flow )
    {
      case "horizontal-wrap": 
      case "horizontal-flow": horz = true; pos = findPosWrap(vx,vy,false); break;
      case "horizontal":      horz = true; pos = findPosHorz(vx,vy); break;
      case "table-body": 
      case "vertical-wrap": 
      case "vertical-flow":   horz = false; pos = findPosWrap(vx,vy,true); break;
      default:                horz = false; pos = findPosVert(vx,vy); break; 
    }
    
    // check for positions that are not allowed in DD:
    if( typeof pos != #integer )
      return;
    else if( pos >= dd_target.length ) // after last pos
    {
      var tc = dd_target.last;
      if(tc === dd_source) tc = tc.prior;
      if(tc && tc.$is(.placeholder))
        return; // not allowing to insert next to placeholder
    }
    else
    {
      var tc = dd_target[pos];
      //if( tc.$is(.placeholder) || (tc.prior && tc.prior.$is(.placeholder))) 
      if( tc.$is(.placeholder)) 
        return;
    }
    // finally setup it:
    if( dd_source.parent === dd_target ) // if elements is moved inside its continer
      setupSrcPlaceholderAt(pos, horz);
    else
      setupDstPlaceholderAt(pos, horz);
  }
  
  function easeOutQuad( t, b, c, d) { return -c *(t/=d)*(t-2) + b; }
  
  function moveIt(what, where, whenDone)
  {
    var easef = def.easeDrop || easeOutQuad;
    if( !easef )
      { whenDone(what,where); return; } // just return
    
    //stdout.println(where.parent);
    
    // do requested animation:  
    var (fromx,fromy,fromw,fromh) = what.box(#rectw, #inner, #view);
    var (tox,toy,tow,toh) = where.box(#rectw, #inner, #view);
    //var progress = 0.0;
    function anim(progress) 
    {
      if( !dd_source || progress >= 1.0 ) { 
        is_animating = false; 
        what.move(); 
        whenDone(what,where); 
        return false; 
      }
      var x = easef(progress, fromx, tox - fromx, 1.0).toInteger();
      var y = easef(progress, fromy, toy - fromy, 1.0).toInteger();
      var w = easef(progress, fromw, tow - fromw, 1.0).toInteger();
      var h = easef(progress, fromh, toh - fromh, 1.0).toInteger();
      what.move(x,y,w,h);
      return true;
    }
    is_animating = true; 
    what.animate(anim, ANI_DURATION);
  }
 
  doDrop = function() 
  {
    assert dd_source && dd_target;
    var dst = dd_placeholder_dst || dd_placeholder_src;
    if(!def.acceptDrop || def.acceptDrop( dd_source, dd_target, dst.index ))
    {
      // OK to drop it here, do it:
      moveIt(dd_source, dst, function()
      {
        var idx = dst.index; 
        
        if( dd_source ) 
        {
          dd_source.move();
          var from = dd_source.parent;
          dd_target.insert(dd_source, idx); // insert our element in place of dd_placeholder_dst
          
          if(dd_placeholder_dst)
          {
            dd_placeholder_dst.remove(); // delete it from the DOM
            dd_placeholder_dst = null;
          }
          
          // unsubscribe if needed
          //if( !dd_source.selectParent(def.what) )
          //  dd_source.unsubscribe(onMouseHandler);
          
          if( dd_placeholder_src ) 
          {
            if(dd_op == #moving) 
              dd_placeholder_src.remove();
            else if(dd_op == #copying) 
            { 
              // cvt our placeholder to normal moveable thing; 
              dd_placeholder_src.@.removeClass("placeholder","src"); 
              //dd_placeholder_src.subscribe(onMouseHandler, Event.MOUSE);
            }
            dd_placeholder_src.style.clear();
            dd_placeholder_src = null;
          }
          if(def.dropped) def.dropped(dd_source, from);
        }
        
        ddFinalize();
      });
    }
    else 
      doCancelDrop();
  }
  doCancelDrop = function() 
  {
    moveIt(dd_source, dd_placeholder_src, function()
    {
      if( dd_source ) {
        dd_source.swap(dd_placeholder_src);
        dd_placeholder_src.remove(); // delete it from the DOM
        dd_placeholder_src = null;
        dd_source.move();
      }
      ddFinalize();
    });
  }

  setupDstPlaceholderAt = function(pos, horz) 
  {
    if(!dd_placeholder_dst) // if there was no dd_placeholder_dst before create it:
    {
      dd_placeholder_dst = new Element(dd_source.tag);
      dd_placeholder_dst.@#class = "placeholder dst";
      dd_target.insert(dd_placeholder_dst,pos);
      if(horz) dd_placeholder_dst.style#width = px(dd_source.box(#width,#inner,#self));
      else     dd_placeholder_dst.style#height = px(dd_source.box(#height,#inner,#self));
    }
    else
      dd_target.insert(dd_placeholder_dst,pos);

    view.update();      
  }
  
  setupSrcPlaceholderAt = function(pos, horz) 
  {
    dd_target.insert(dd_placeholder_src,pos); // just move dd_placeholder_src here

    view.update();
    // just in case it is inside scrollable container make 
    // next/previous element visible
    if(dd_autoScroll) {
      if(dd_placeholder_src.prior)
        dd_placeholder_src.prior.scrollToView(false,false);
      if(dd_placeholder_src.next)
        dd_placeholder_src.next.scrollToView(false,false);
    }
  }
  
  function offset( parent, child )
  {
    var (px,py) = parent.box(#position, #inner, #view);
    var (cx,cy) = child.box(#position, #inner, #view);
    return (cx - px, cy - py);
  }
  
  function localCoord(el,evt) {
    var (tx,ty) = el.box(#position, #inner, #view);
    tx = evt.xView - tx;
    ty = evt.yView - ty;
    return (tx,ty);
  }
  
  var xViewPos, yViewPos;
  
  function ping() {
    var el = view.root.find(xViewPos, yViewPos);
    if( el )
      el.postEvent("drag-n-drop-ping"); // generate "ping" event in case of UI need to scrolling, etc.
  }
  
  function draggableMouseHandler(evt)
  { 
    switch(evt.type)  
    {
      case Event.MOUSE_DOWN | Event.SINKING:  
      {
        dd_movable_el = this;
        (dd_x,dd_y) = localCoord(this,evt);
        return false;
      }
      case Event.MOUSE_UP | Event.SINKING:  
        dd_x = dd_y = dd_movable_el =null; 
        this.timer(0, ping);
        return true; 
      case Event.MOUSE_ENTER | Event.SINKING:  
      case Event.MOUSE_LEAVE | Event.SINKING:  
        if(!dd_source) {
          dd_x = dd_y = dd_movable_el =null; 
        }
        break; 
        
      //case Event.MOUSE_TICK | Event.SINKING:  
      //  stdout.println("Event.MOUSE_TICK");
      //  break;
  
      case Event.MOUSE_MOVE | Event.SINKING: 
        if( !evt.mainButton ) 
          return; 
        if( is_animating )
          return;
        if(dd_source) {
          xViewPos = evt.xView;
          yViewPos = evt.yView;
          this.timer(PING_THRESHOLD, ping);
          return doMove( xViewPos, yViewPos);
        }
        else if( typeof dd_x == #integer )
        {
          var (x,y) = localCoord(this,evt);
          var deltax = dd_x - x;
          var deltay = dd_y - y;
          //stdout.$n({dd_x} {dd_y});
          if( deltax < -X_THRESHOLD || deltax > X_THRESHOLD ||
              deltay < -Y_THRESHOLD || deltay > Y_THRESHOLD )
              dd_x = x;
              dd_y = y;
              doDD( this, evt.xView, evt.yView );
          return true;
        }
    }
  }
  // ready to go, attach onMouseHandler to the draggables
  
  function validDraggable(draggable) {
    for(var t in dd_targets)
      if( draggable.belongsTo(t,true,true) )
              return true;
    return false;
  }

  function mouseEventMonitor(evt) {
    if( !evt.target )
      return false;

    var draggable = evt.target.selectParent(def.what);
    if( draggable && validDraggable(draggable) ) {
      if (dd_movable) {
        draggable = draggable.selectParent(dd_movable);
      } 
      return draggableMouseHandler.call(draggable,evt);
    } else {
      if (dd_movable_el && evt.target == dd_movable_el) {
        draggable = evt.target;
        return draggableMouseHandler.call(draggable,evt);
      }

    }
  }

  function ddCancel() { // cancel DD loop
    requested_cancel = true;
    if(dd_dragging) 
      dd_dragging.capture(false); // remove capture, stop view.doEvent(#untilMouseUp) loop
  }

  function ddShutdown() { // cancel DD loop and remove traces of this DragDrop call.
    ddCancel(); 
    self.select(def.container).unsubscribe(mouseEventMonitor); 
  }

  self.select(def.container).subscribe(mouseEventMonitor, Event.MOUSE); 
  
  return {
    cancel : ddCancel, 
    remove : ddShutdown
  };
  
  
}

</script>

    <!-- res/animations.tis -->
 <script type="text/tiscript">
    namespace Animation
    {
      const animatableExpandAtts = [#height,#width,
                                    #border-top-width, #border-bottom-width, #border-left-width, #border-right-width,
                                    #border-top-color, #border-bottom-color, #border-left-color, #border-right-color,
                                    #padding-top, #padding-bottom, #padding-left, #padding-right,
                                    #margin-top,#margin-bottom,#margin-left,#margin-right,
                                    #background-color-top-left,#background-color-top-right, #background-color-bottom-left,#background-color-bottom-right,
                                    #opacity ]; 
      const animatableMoveAtts =   [#height,#width,
                                    #top, #bottom, #left, #right ]; 

      namespace Ease // collection of easing functions of Robert Penner
      {      
        // See: http://www.robertpenner.com/easing/
        
        // signature of methods: (current_time,	start_value, end_value-start_value,	total_time)

        function Linear( t, b, c, d) 
          {
            return (t/d)*c + b;
          }
        function InQuad( t, b, c, d) 
          {
            return c*(t/=d)*t + b;
          }
	      function OutQuad( t, b, c, d) 
          {
            return -c *(t/=d)*(t-2) + b;
          }
        function InOutQuad( t, b, c, d) 
          {
            if ((t/=d/2) < 1) return c/2*t*t + b;
            return -c/2 * ((--t)*(t-2) - 1) + b;
          }
        function InCubic( t, b, c, d) 
          {
            return c*(t/=d)*t*t + b;
          }
        function OutCubic( t, b, c, d) 
          {
            return c*((t=t/d-1)*t*t + 1) + b;
          }
        function InOutCubic( t, b, c, d) 
          {
            if ((t/=d/2) < 1) return c/2*t*t*t + b;
            return c/2*((t-=2)*t*t + 2) + b;
          }
	      function InQuart( t, b, c, d) 
          {
            return c*(t/=d)*t*t*t + b;
          }
	      function OutQuart ( t, b, c, d) 
          {
            return -c * ((t=t/d-1)*t*t*t - 1) + b;
          }
        function InOutQuart ( t, b, c, d) 
          {
            if ((t/=d/2) < 1) return c/2*t*t*t*t + b;
            return -c/2 * ((t-=2)*t*t*t - 2) + b;
          }
        function InQuint ( t, b, c, d) 
          {
            return c*(t/=d)*t*t*t*t + b;
          }
        function OutQuint ( t, b, c, d) 
          {
            return c*((t=t/d-1)*t*t*t*t + 1) + b;
          }
        function InOutQuint( t, b, c, d) 
          {
            if ((t/=d/2) < 1) return c/2*t*t*t*t*t + b;
            return c/2*((t-=2)*t*t*t*t + 2) + b;
          }
        function InSine( t, b, c, d) 
          {
            return -c * Math.cos(t/d * (Math.PI/2)) + c + b;
          }
        function OutSine( t, b, c, d) 
          {
            return c * Math.sin(t/d * (Math.PI/2)) + b;
          }
	      function InOutSine( t, b, c, d) 
          {
            return -c/2 * (Math.cos(Math.PI*t/d) - 1) + b;
          }
        function InExpo( t, b, c, d) 
          {
            return (t==0) ? b : c * Math.pow(2, 10 * (t/d - 1)) + b;
          }
        function OutExpo( t, b, c, d) 
          {
            return (t==d) ? b+c : c * (-Math.pow(2, -10 * t/d) + 1) + b;
          }
        function InOutExpo( t, b, c, d) 
          {
            if (t==0) return b;
            if (t==d) return b+c;
            if ((t/=d/2) < 1) return c/2 * Math.pow(2, 10 * (t - 1)) + b;
            return c/2 * (-Math.pow(2, -10 * --t) + 2) + b;
          }
        function InCirc( t, b, c, d) 
          {
            return -c * (Math.sqrt(1 - (t/=d)*t) - 1) + b;
          }
        function OutCirc( t, b, c, d) 
          {
            return c * Math.sqrt(1 - (t=t/d-1)*t) + b;
          }
        function InOutCirc( t, b, c, d) 
          {
            if ((t/=d/2) < 1) return -c/2 * (Math.sqrt(1 - t*t) - 1) + b;
            return c/2 * (Math.sqrt(1 - (t-=2)*t) + 1) + b;
          }
        function InElastic( t, b, c, d) 
          {
            var s=1.70158;var p=0;var a=c;
            if (t==0) return b;  if ((t/=d)==1) return b+c;  if (!p) p=d*.3;
            if (a < Math.abs(c)) { a=c; s=p/4; }
            else s = p/(2*Math.PI) * Math.asin (c/a);
            return -(a*Math.pow(2,10*(t-=1)) * Math.sin( (t*d-s)*(2*Math.PI)/p )) + b;
          }
        function OutElastic( t, b, c, d) 
          {
            var s=1.70158;var p=0;var a=c;
            if (t==0) return b;  if ((t/=d)==1) return b+c;  if (!p) p=d*.3;
            if (a < Math.abs(c)) { a=c; s=p/4; }
            else s = p/(2*Math.PI) * Math.asin (c/a);
            return a*Math.pow(2,-10*t) * Math.sin( (t*d-s)*(2*Math.PI)/p ) + c + b;
          }
        function InOutElastic( t, b, c, d) 
          {
            var s=1.70158;var p=0;var a=c;
            if (t==0) return b;  if ((t/=d/2)==2) return b+c;  if (!p) p=d*(.3*1.5);
            if (a < Math.abs(c)) { a=c; s=p/4; }
            else s = p/(2*Math.PI) * Math.asin (c/a);
            if (t < 1) return -.5*(a*Math.pow(2,10*(t-=1)) * Math.sin( (t*d-s)*(2*Math.PI)/p )) + b;
            return a*Math.pow(2,-10*(t-=1)) * Math.sin( (t*d-s)*(2*Math.PI)/p )*.5 + c + b;
          }
        function InBack( t, b, c, d, s = 1.70158) 
          {
            return c*(t/=d)*t*((s+1)*t - s) + b;
          }
        function OutBack( t, b, c, d, s = 1.70158) 
          {
            return c*((t=t/d-1)*t*((s+1)*t + s) + 1) + b;
          }
        function InOutBack( t, b, c, d, s = 1.70158) 
          {
            if ((t/=d/2) < 1) return c/2*(t*t*(((s*=(1.525))+1)*t - s)) + b;
            return c/2*((t-=2)*t*(((s*=(1.525))+1)*t + s) + 2) + b;
          }
        function OutBounce( t, b, c, d) 
          {
            if ((t/=d) < (1/2.75)) 
              return c*(7.5625*t*t) + b;
            else if (t < (2/2.75))
              return c*(7.5625*(t-=(1.5/2.75))*t + .75) + b;
            else if (t < (2.5/2.75)) 
              return c*(7.5625*(t-=(2.25/2.75))*t + .9375) + b;
            else 
              return c*(7.5625*(t-=(2.625/2.75))*t + .984375) + b;
          }
        function InBounce( t, b, c, d) 
          {
            return c - OutBounce ( d-t, 0, c, d) + b;
          }
        function InOutBounce( t, b, c, d) 
          {
            if (t < d/2) return InBounce ( t*2, 0, c, d ) * .5 + b;
            return OutBounce ( t*2-d, 0, c, d ) * .5 + c*.5 + b;
          }
      } //namespace Ease 

                              
      function morphStyle(element, styleAtts, stepDelay, switchStateFunction, easeF, clearAfter = true)
      {
        // makes snapshot of CSS atts we are interesting in
        function styleSnapshot(element)
        {
          var col = {};
          for(var styleAtt in styleAtts)  
          {
            var v = element.style[styleAtt];
            if( !v ) continue;
            col[styleAtt] = v;
          }
          return col;
        }
        
        var progress = 0.0; 
               
        // makes mid value between initVal and finalVal according to progress value (0..100)
        function makeMidValue(initVal, finalVal)
        {
          if( typeof initVal == #color &&  typeof finalVal == #color) // color value
          {
            // easing is linear for colors
            var r = (progress * (finalVal.r - initVal.r) + initVal.r).toInteger();
            var g = (progress * (finalVal.g - initVal.g) + initVal.g).toInteger();
            var b = (progress * (finalVal.b - initVal.b) + initVal.b).toInteger();
            return color(r,g,b);
          }
          if( typeof initVal == #length ) // length value
          {
            var units = initVal.units;
            initVal = initVal.toFloat();
            finalVal = finalVal.toFloat();
            var v = easeF(progress, initVal, finalVal - initVal, 1.0);
            ////var v = progress * (finalVal.toFloat() - initVal.toFloat()) / 100 + initVal.toFloat();
            //stdout.printf("v=%v progress=%v %v %v\n",v,progress,initVal.toFloat(), finalVal.toFloat());
            return length(v, units );
          }
        }
        
        var initialStyles = styleSnapshot(element);
        switchStateFunction();
        var finalStyles = styleSnapshot(element);
        
        function do_morph()
        {
          progress += 0.02;
          if( progress > 1.0 ) 
          {
            if(clearAfter)
              element.style.clear(); // clear runtime styles we set in animation
            return;
          }
          for( var satt in initialStyles )
          {
            var initVal = initialStyles[satt];
            var finalVal = finalStyles[satt];
            if( !initVal || !finalVal )
              continue;
            if( initVal == finalVal )
              continue; // nothing to do
            //stdout << satt << " " << initVal << " " << finalVal << "\n";
            element.style[satt] = makeMidValue(initVal, finalVal);
          }
          return stepDelay;
        }
        element.animate(do_morph);
      }
      
      function expand(element, easeF = Animation.Ease.InQuad, stepDelay = 8)
      {
        function toggle()
        {
          if(element.state.expanded)
            element.state.collapsed = true;
          else
            element.state.expanded = true;
        }
        morphStyle(element, animatableExpandAtts, stepDelay, toggle, easeF);
      }
      
      function move( element, x, y, easeF = Animation.Ease.InQuad, stepDelay = 6)
      {
        function toggle()
        {
          element.style#left = px( x );
          element.style#top = px( y );
        }
        morphStyle(element, animatableMoveAtts, stepDelay, toggle, easeF, false);
      }
      


    }  

</script>

    <!-- res/app.tis -->
 <script type="text/tiscript">
self.ready = function() {
	App.init();
}


namespace App
{  
	const TAG = "#@GOHOSTS ";
	const SWITCHHOST_TAG = "#@SwitchHosts! ";
	const SEPERATE = "# --------------------------------------------------";
	const TITLE_MAX_LENGTH = 18;
	var hosts = [{icon:#system,title:"System Hosts", name:"system", active:true, on: false, timer: null, dirty: false}]; // observable variable 
	var activeHostIndex = 0;
	var activeHostName = "system";
	var editHostTitle = "";
	var editHostIndex = -1;
	var searchItemsCount = -1;
	var config = {};
	var hosticon = "&fa-desktop;";

	function init() {
		view.isMaximizable = false;
		view.mkdir(System.home("hosts"));

		loadConfig();

		var onMeta = null;
		try {
		    var content = Bytes.load(getSystemHostPath()).toString("UTF-8");
		    onMeta = getFileMeta(content);
		    $(plaintext).value = removeFileMeta(content);
		    $(plaintext).state.readonly = true;
	    } catch(e) {
	    	// host file need authorization for user: Authenticated User
	    	view.msgbox(#warning, "Cannot read hosts file: " + System.path(#SYSTEM_BIN, "drivers/etc/hosts") );
	    }

	    getHostFiles(onMeta);

	    // double click to toggle comment
	    $(plaintext).on("dblclick", function(evt) {
	    	if (evt.x <= 38) {
	    		var line = evt.target;
	    		toggleComment(line);
	    	}
	    	return true;
		});

		// edit host
		$(#edit-form input).on("keypress", function(evt) {
			if (evt.keyCode == 13) {  // enter
				doEditHost();
				return false;
			}
		});

	    // ctrl+f to search text
		$(plaintext).on("keypress", function(evt) {
			
		    if( evt.ctrlKey && evt.keyCode == 6) { // ctrl+f
		    	$(#searchbox).attributes.removeClass("hide");
		    	$(#searchbox>input).state.focus = true;
		    	$(#searchbox>input).execCommand("edit:selectall");
		    	return true;
		    } else if( evt.ctrlKey && evt.keyCode == 3) { // ctrl+c
		    	var text = view.clipboard(#get, #text);
		    	if (text) {
			    	view.clipboard(#put, formatContentSpace(text));
			    }
			    return true;
		    } else if (evt.keyCode == Event.VK_ESCAPE) { // esc
		    	hideSearchBox();
		    	return true;
		    }
		});
		$(plaintext).on("exec:edit:paste", function() {
			// bug hack!!!
			// when clipboard text exceed 65535 chars, sciter clipboard will cut off more than 65535 chars
			// here use golang win32 function to get all clipboard text and append
			var text = view.clipboardText();
	    	if (text && text.length > 65000) {
	    		var startIndex = $(plaintext).selection.start[0].parent.index;
	    		var endIndex = $(plaintext).selection.end[0].parent.index;
	    		var selectionLen = $(plaintext).selection.text.length;
	    		var selectionStartIndex = $(plaintext).selection.start[1];
	    		var selectionEndIndex = $(plaintext).selection.end[1];
	    		$(plaintext).selection.collapse(#toCaret);

	    		var content = "";
	    		for (var i = 0; i < $(plaintext).length; i++) {
	    			if (i >= startIndex && i <= endIndex) {
	    				if (startIndex == endIndex) {
	    					content += $(plaintext)[i].text.splice(selectionStartIndex, selectionLen, text) + "\r\n";
	    				} else {
	    					if (i == startIndex) {
	    						content += $(plaintext)[i].text.splice(selectionStartIndex, $(plaintext)[i].text.length, text);
	    					} else if (i == endIndex) {
	    						content += $(plaintext)[i].text.splice(0, selectionEndIndex + 1) + "\r\n";
	    					}
	    				}
	    			} else {
			    		content += $(plaintext)[i].text + "\r\n";
			    	}
		    	}
		    	$(plaintext).value = content;
		    	$(plaintext).sendEvent("change");
		    	return true;
		    }
        });
		$(#searchbox>input).on("keypress", function(evt) {
		    if (evt.keyCode == Event.VK_ESCAPE) { // esc
		    	hideSearchBox();
		    	return true;
		    } else if (evt.keyCode == 13) {  // enter
		    	var findText = this.value;

		    	var allMatchBookMark = [];
		    	for (var i = 0; i < $(plaintext).length; i++) {
		    		var text = $(plaintext)[i];
		    		var idx = text.value.indexOf(findText);
		    		if (idx >= 0) {
		    			var start = [bookmark: text.firstNode, idx, false];
		    			var end = [bookmark: text.firstNode, idx+findText.length-1, true];
		    			allMatchBookMark.push({index: i, start:start, end:end})
		    		}
		    	}

		    	searchItemsCount = allMatchBookMark.length;

		    	if (allMatchBookMark.length > 0) {
		    		var startLine = $(plaintext).selection.start[0].parent.index;
		    		if (startLine >= allMatchBookMark[allMatchBookMark.length -1].index) {
		    			$(plaintext).selection.select(allMatchBookMark[0].end, allMatchBookMark[0].start);
		    		} else {
		    			for (var i = 0; i < allMatchBookMark.length; i++) {
		    				if (allMatchBookMark[i].index > startLine) {
		    					$(plaintext).selection.select(allMatchBookMark[i].end, allMatchBookMark[i].start);
		    					break;
		    				}
		    			}
		    		}
		    	}
		    	
		    	return true;
		    }
		});


		$(menu#context).on("click", "li#context_edit_host", function(evt) { 
			var name = $(#menu li:hover).attributes["data-name"];
			for (var i = 0; i < hosts.length; i++) {
				if (i == 0) {
					continue;
				}

				if (hosts[i].name == name) {
					editHostTitle = hosts[i].title;
					editHostIndex = i;

					editHost();
					break;
				}
			}

			return true;
		});
		$(menu#context).on("click", "li#context_delete_host", function(evt) { 
		    var name = $(#menu li:hover).attributes["data-name"];
			for (var i = 0; i < hosts.length; i++) {
				if (i == 0) {
					continue;
				}

				if (hosts[i].name == name) {
					hosts[i].active = true;
					deleteHost();
					break;
				}
			}

			return true;
		});


		DragDrop
	    {
	      what      : "#menu ul > li > span.icon",
	      notBefore : "#menu ul > li:first-child",
	      movable   : "#menu ul > li",
	      where     : "#menu ul",
	      container : "#menu ul",
	      dropped: function(draggable, from) {
	      	saveHostListSort();
	      },
	      easeDrop  : Animation.Ease.OutQuad
	    };
	}

	function getHostFiles(onMeta) {
		var tmpHosts = {}
		System.scanFiles(getHostPath("*.*"), function(filename, attrs) {
			if(attrs & (System.IS_DIR | System.IS_HIDDEN)) return true;

			var title = filename;
			var path = getHostPath(filename);
			var content = Bytes.load(path).toString("UTF-8");
			var meta = getFileMeta(content);
			if (meta) {
				title = meta.title;
			}
			if (title.length > TITLE_MAX_LENGTH) {
				title = title.substr(0, TITLE_MAX_LENGTH) + "...";
			}

			tmpHosts[filename] = {icon:#host, title:title, name:filename, active:false, on: onMeta && onMeta.name == filename, timer: null, dirty: false};
			return true;
		});

		if (config.sort) {
			for (var i = 0; i < config.sort.length; i++) {
				var name = config.sort[i];
				if (tmpHosts[name]) {
					hosts.push(tmpHosts[name]);
					delete tmpHosts[name];
				}
			}
		}

		for (var name in tmpHosts) {
			if (name) {
				hosts.push(tmpHosts[name]);
			}
		}
	}


	function showHost(host) {
		if (host.name == activeHostName) {
			return;
		}

		doSave(true);


		var path = getHostPath(host.name);
		if (host.name == "system") {
			path = getSystemHostPath();
			$(plaintext).state.readonly = true;
		} else {
			$(plaintext).state.readonly = false;
		}
		var content = getFileContent(path);
		$(plaintext).value = content;


		toggleActiveHost(host.name);
	}


	function switchHost(host) {
		if (host.name == "system") {
			return;
		}

		doSave(true);

		// stdout.println("double click");
		var path = getHostPath(host.name);
		var content = getFileContent(path);

		var meta = TAG + JSON.stringify({title: host.title, name: host.name});
		content = meta + "\r\n\r\n" + SEPERATE + "\r\n" + formatContentSpace(content);

		var success = Bytes.fromString(content, "UTF-8").save(getSystemHostPath());
		if (success) {
			toggleOnHost(host.name);
			view.clearDNSCache();
			Notification.show("Switch host success.", "Notification");  
		} else {
			view.msgbox(#warning, "Save host file failed!");
		}
	}

	function formatContentSpace(content) {
		if (!content) {
			return "";
		}

		// when mix /t and space in content, sometimes
		// windows 7 will ignore it
		return content.replace(/\r\n/g, "[n]")
		              .replace(/\n/g, "[n]")
		              .replace(/\t/g, "    ")
		              .replace(/\s/g, " ")
		              .replace(/\[n\]/g, "\r\n");
	}

	function editHost() {
		if (editHostIndex < 0) {
			$(#edit-form input).state.focus = true;
		} else {
			//$(#edit-form input).setSelection(editHostTitle.length, editHostTitle.length);
			$(#edit-form h2).attributes.addClass("edit");
		}
		$(#overlay).attributes.toggleClass("hide", false);
		$(#edit-form).attributes.toggleClass("hide", false);
	}

	function cancelEdit() {
		editHostTitle = "";
		editHostIndex = -1;
		$(#msg).text = "";
		$(#msg).attributes.toggleClass("hide", true);
		$(#overlay).attributes.toggleClass("hide", true);
		$(#edit-form).attributes.toggleClass("hide", true);
		$(#edit-form h2).attributes.removeClass("edit");
	}

	function doEditHost() {
		var title = editHostTitle.trim();

		if (title == "") {
			$(#msg).text = "title cannot be empty.";
			$(#msg).attributes.toggleClass("hide", false);
			return;
		}

		for (var i = hosts.length - 1; i >= 0; i--) {
			if (hosts[i].title == title) {
				$(#msg).text = "title has exist.";
				$(#msg).attributes.toggleClass("hide", false);
				return;
			}
		}

		var name = title.replace(/\s+/g, "_");
		if (name == "system") {
			$(#msg).text = "[system] is a reserve word, cannot use.";
			$(#msg).attributes.toggleClass("hide", false);
			return;
		}

		if (editHostIndex > 0) {
			var oldname = hosts[editHostIndex].name;
			var oldPath = getHostPath(oldname);
			var newPath = getHostPath(name);
			view.renameFile(oldPath, newPath);

			hosts[editHostIndex].title = title;
			hosts[editHostIndex].name = name;
			saveFileContent(newPath, getFileContent(newPath), hosts[editHostIndex]);
			saveHostListSort(null, oldname, name);
		} else {
			hosts.push({icon:#host, title:title, name:name, active:false, on: false, timer: null, dirty: false});
			var path = getHostPath(hosts[hosts.length -1].name);
			saveFileContent(path, "#" + title, hosts[hosts.length -1]);
			showHost(hosts[hosts.length -1]);
			saveHostListSort(null, null, name);
		}

		cancelEdit();
	}

	function deleteHost() {
		for (var i=hosts.length-1; i>=0; i--) {
			if (i != 0 && hosts[i].active) {
				var deleteHostName = hosts[i].name;
				view.deleteFile(getHostPath(deleteHostName))
				hosts.remove(i);
				activeHostIndex = 0;
				activeHostName = "system";
				saveHostListSort(deleteHostName);
				break;
			}
		}
	}

	function getFileContent(path) {
		try {
		    return removeFileMeta(Bytes.load(path).toString("UTF-8"));
	    } catch(e) {
	    	// host file need authorization for user: Authenticated User
	    	view.msgbox(#warning, "Cannot read hosts file: " + path );
	    }

	    return "";
	}

	function saveFileContent(path, content, host) {
		var meta = TAG + JSON.stringify({title: host.title, name: host.name});
		content = meta + "\r\n\r\n" + SEPERATE + "\r\n" + content;

		return Bytes.fromString(content, "UTF-8").save(path);
	}

	function getFileMeta(content) {
		var lines = content.split("\n");
		if (lines.length > 0) {
			var firstLine = lines[0];
			
			if (firstLine.substr(0, TAG.length) == TAG) {
				var meta = firstLine.substr(TAG.length).trim();
				return parseData(meta);
			}

			if (firstLine.substr(0, SWITCHHOST_TAG.length) == SWITCHHOST_TAG) {
				var meta = firstLine.substr(SWITCHHOST_TAG.length).trim();
				return parseData(meta);
			}
		}

		return null;
	}


	function removeFileMeta(content) {
		var lines = content.split("\n");
		if (lines.length > 0) {
			var firstLine = lines[0];
			
			if (firstLine.substr(0, TAG.length) == TAG) {
				lines.shift();
			} else if (firstLine.substr(0, SWITCHHOST_TAG.length) == SWITCHHOST_TAG) {
				lines.shift();
			}
		}

		if (lines.length > 0) {
			var brankLine = lines[0];
			if (brankLine.trim() == "") {
				lines.shift();
			}
		}

		if (lines.length > 0) {
			var line = lines[0];
			if (line.substr(0, SEPERATE.length) == SEPERATE) {
				lines.shift();
			}
		}

		return lines.join("\n");
	}




	function toggleActiveHost(name) {
		for (var i=0; i<hosts.length; i++) {
			if (hosts[i].name == name) {
				hosts[i].active = true;
				activeHostIndex = i;
				activeHostName = name;
			} else {
				hosts[i].active = false;
			}
		}
	}

	function toggleOnHost(name) {
		for (var i=0; i<hosts.length; i++) {
			if (hosts[i].name == name) {
				hosts[i].on = true;
			} else {
				hosts[i].on = false;
			}
		}
	}

	function getSystemHostPath() {
		return System.path(#SYSTEM_BIN, "drivers/etc/hosts");
	}

	function getHostPath(name) {
		var dir = System.home("hosts")
		return dir + "/" + name;
	}

	function changeHost() {
		if (activeHostIndex != 0) {
			hosts[activeHostIndex].dirty = true;
			doSave(false);
		}
	}

	function doSave(directUpdate) {
		if (activeHostIndex == 0) {
			return;
		}

		if (!hosts[activeHostIndex].dirty) {
			return;
		}


		var idx = activeHostIndex;
		if (hosts[idx].timer) {
			clearTimeout(hosts[idx].timer);
		}
		var path = getHostPath(hosts[idx].name);
		var content = $(plaintext).value;
		var host = hosts[idx];

		if (directUpdate) {
			saveFileContent(path, content, host);
			hosts[idx].timer = null;
			hosts[idx].dirty = false;
		} else {
			// delay 1 seconds to save input
			hosts[idx].timer = setTimeout(function() {
				// stdout.println("timeout");
				saveFileContent(path, content, host);
				hosts[idx].timer = null;
				hosts[idx].dirty = false;
			}, 1000);
		}
	}

	function setTimeout( func, milliseconds ) {
	   function timerCallback(  ) { func(); return false; }
	   self.timer(milliseconds, timerCallback, true);
	   return timerCallback; // returns function reference as unique interval id.
	}

	function clearTimeout( id ) {
	   self.timer(0, id, true);
	}

	function toggleComment(line) {
		if (!line.value) {
			return;
		}

		if (line.value.match(/^\s*?#/)) {
			line.value = line.value.replace(/^\s*?#/, "");
		} else {
			line.value = "#" + line.value;
		}
		$(plaintext).sendEvent("change");
	}

	function hideSearchBox() {
		searchItemsCount = -1;

		if (!$(#searchbox).attributes.hasClass("hide")) {
	    	$(#searchbox).attributes.addClass("hide");
	    }
	    if (!$(plaintext).focus) {
	    	$(plaintext).focus = true;
	    }
	}

	function saveHostListSort(deleteHostName = null , oldHostName = null, newHostName = null) {
		var sortContent = [];
      	for(var li in $(#menu>ul)) {
      		if (li.attributes["data-name"] && li.attributes["data-name"] != "system") {
      			var hostname = li.attributes["data-name"];
      			if (deleteHostName && deleteHostName == hostname) {
      				continue;
      			}
      			if (oldHostName && newHostName && oldHostName == hostname) {
      				hostname = newHostName;
      			}
	      		sortContent.push(hostname);
	      	}
      	}
      	if (!oldHostName && newHostName) {
      		sortContent.push(newHostName);
      	}
      	config.sort = sortContent;
      	saveConfig();
	}

	function loadConfig() {
		try {
			var path = System.home("config.json");
			if (view.fileExists(path)) {
				var content = Bytes.load(System.home("config.json")).toString("UTF-8")
				config = JSON.parse(content);
			}
		} catch(e) {
			view.msgbox(#warning, e);
		}
	}

	function saveConfig() {
		if (!config) {
			return;
		}
		try {
			var content = JSON.stringify(config);
			Bytes.fromString(content, "UTF-8").save(System.home("config.json"));
		} catch(e) {
			view.msgbox(#warning, e);
		}
	}


}  


</script>

</body>

</html>
`
