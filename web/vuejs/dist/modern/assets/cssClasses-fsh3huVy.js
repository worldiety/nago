import{r as n}from"./util-Bvzg2mFt.js";class o{static classes=new Map;static defaultStyleElem;static hoverStyleElem;static focusStyleElem;static activeStyleElem;static init(){this.defaultStyleElem&&this.hoverStyleElem&&this.focusStyleElem&&this.activeStyleElem||(this.defaultStyleElem=document.createElement("style"),this.hoverStyleElem=document.createElement("style"),this.focusStyleElem=document.createElement("style"),this.activeStyleElem=document.createElement("style"),document.body.appendChild(this.defaultStyleElem),document.body.appendChild(this.hoverStyleElem),document.body.appendChild(this.focusStyleElem),document.body.appendChild(this.activeStyleElem))}static getOrCreate(e,t){const s=this.createMapKey(e,t),l=this.classes.get(s);return l||this.createCssClass(s,e,t)}static createCssClass(e,t,s){this.init();const l=t.join(`;
`),i=n(12);let a=`.${i}${s?`:${s}`:""}`;s==="focus"&&(a+=`,
.${i}:focus-visible`);const c=`${a} {
${l}
}`;return this.addClassToElem(c,s),this.classes.set(e,i),i}static addClassToElem(e,t){switch(t){case"hover":this.hoverStyleElem.innerHTML+=`

${e}`;break;case"focus":this.focusStyleElem.innerHTML+=`

${e}`;break;case"active":this.activeStyleElem.innerHTML+=`

${e}`;break;default:this.defaultStyleElem.innerHTML+=`

${e}`}}static createMapKey(e,t){return`${e.join("-")}-${t||"default"}`}}export{o as C};
