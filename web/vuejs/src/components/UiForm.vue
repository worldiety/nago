<script lang="ts" setup>
import type {Form, FormField} from '@/shared/model';
import {ref} from "vue";
import UiGeneric from "@/components/UiGeneric.vue";
import router from "@/router";
import {useHttp, userHeaders} from "@/shared/http";

const props = defineProps<{
  ui: Form;
}>();

const formFields = ref<FormField[]>(new Array<FormField>())

const http = useHttp();
const headers = userHeaders();


async function init(): Promise<void> {
  if (props.ui.links.load != null) {
    const resp = await http.request(props.ui.links.load).then((r) => r.json())
    formFields.value = resp['fields']
  }else{
    console.log("warning: form has no load function defined, this is not allowed")
  }

}

async function sendAllForms(isDelete:boolean): Promise<void> {

  const formData = new FormData();
  const inputElems = document.getElementsByTagName('input');
  for (let i = 0; i < inputElems.length; i++) {
    const item = inputElems.item(i);
    if (item == null) { // ts and linters are so stupid...
      throw new Error('cannot happen!?');
    }


    const name = item.getAttribute('name');
    if (name == null || name == '') {
      continue;
    }



    if (item.getAttribute('type') === 'file') {
      if (item.files == null) {
        continue;
      }

      for (const file of item.files) {
        formData.append(name, file);
      }
    } else {

      if (item.getAttribute('type') === 'checkbox'){
        formData.append(name,item.checked)
      }else{
        formData.append(name,item.value)
      }

    }
  }


  const textAreaElems = document.getElementsByTagName('textarea');
  for (let i = 0; i < textAreaElems.length; i++) {
    const item = textAreaElems.item(i);
    if (item == null) { // ts and linters are so stupid...
      throw new Error('cannot happen!?');
    }


    const name = item.getAttribute('name');
    if (name == null || name == '') {
      continue;
    }
    formData.append(name,item.value)

  }

  if (isDelete){
    formData.append("_action","delete")
  }else{
    formData.append("_action","update")
  }

  console.log(formData)

  const h = await headers.headers()
  const uploadRes = await fetch(props.ui.links.submit!, {method:'POST',body: formData,headers:h}).then((r)=>r.json())
  console.log(uploadRes)

  if (uploadRes.type === "FormValidationError"){
    formFields.value = uploadRes['fields']
    return
  }

  if (uploadRes.type ==="Redirect"){
    router.push(uploadRes.target)
    return
  }
}

init();
console.log("UiForm init")
</script>

<template>


  <ui-generic v-for="field in formFields" :ui="field"/>
  <v-responsive
      class="mx-auto"
      max-width="344"
  >
    <v-btn v-if="props.ui.links.submit"
           class="me-4 mt-2"
           block
           @click="sendAllForms(false)"
    >
      {{ props.ui.submitText }}
    </v-btn>

    <v-btn v-if="props.ui.links.delete"
           class="me-4 mt-2"
           block
           color="red"
           @click="sendAllForms(true)"
    >
      {{ props.ui.deleteText }}
    </v-btn>
  </v-responsive>
</template>
