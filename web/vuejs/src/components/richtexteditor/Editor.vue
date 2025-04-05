<script>
import { Color } from '@tiptap/extension-color';
import ListItem from '@tiptap/extension-list-item';
import TextStyle from '@tiptap/extension-text-style';
import { Underline } from '@tiptap/extension-underline';
import StarterKit from '@tiptap/starter-kit';
import { Editor, EditorContent } from '@tiptap/vue-3';

export default {
	components: {
		EditorContent,
	},

	props: {
		modelValue: {
			type: String,
			default: '',
		},
	},

	emits: ['update:modelValue'],

	data() {
		return {
			editor: null,
		};
	},

	watch: {
		modelValue(value) {
			// HTML
			const isSame = this.editor.getHTML() === value;

			// JSON
			// const isSame = JSON.stringify(this.editor.getJSON()) === JSON.stringify(value)

			if (isSame) {
				return;
			}

			this.editor.commands.setContent(value, false);
		},
	},

	mounted() {
		this.editor = new Editor({
			extensions: [
				Color.configure({ types: [TextStyle.name, ListItem.name] }),
				TextStyle.configure({ types: [ListItem.name] }),
				StarterKit,
				Underline,
			],
			content: this.modelValue,
			onUpdate: () => {
				// HTML
				this.$emit('update:modelValue', this.editor.getHTML());

				// JSON
				// this.$emit('update:modelValue', this.editor.getJSON())
			},
			onBlur: (props) => {
				this.$emit('blur', event);
			},
		});
	},

	beforeUnmount() {
		this.editor.destroy();
	},
};
</script>

<template>
	<div v-if="editor" class="container">
		<div class="control-group">
			<div class="gap-1 flex flex-wrap">
				<button
					@click="editor.chain().focus().toggleBold().run()"
					:disabled="!editor.can().chain().focus().toggleBold().run()"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('bold') }"
					title="Bold"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M8 5h4.5a3.5 3.5 0 1 1 0 7H8m0-7v7m0-7H6m2 7h6.5a3.5 3.5 0 1 1 0 7H8m0-7v7m0 0H6"
						/>
					</svg>
				</button>
				<button
					@click="editor.chain().focus().toggleItalic().run()"
					:disabled="!editor.can().chain().focus().toggleItalic().run()"
					title="italic"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('italic') }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="m8.874 19 6.143-14M6 19h6.33m-.66-14H18"
						/>
					</svg>
				</button>

				<button
					@click="editor.chain().focus().toggleUnderline().run()"
					:disabled="!editor.can().chain().focus().toggleUnderline().run()"
					title="underline"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('underline') }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-width="2"
							d="M5 19h14M7.6 16l4.2979-10.92963c.0368-.09379.1674-.09379.2042 0L16.4 16m-8.8 0H6.5m1.1 0h1.65m7.15 0h-1.65m1.65 0h1.1m-8.33315-4h5.66025"
						/>
					</svg>
				</button>

				<button
					@click="editor.chain().focus().toggleStrike().run()"
					:disabled="!editor.can().chain().focus().toggleStrike().run()"
					title="strike"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('strike') }"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-6"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M12 12a8.912 8.912 0 0 1-.318-.079c-1.585-.424-2.904-1.247-3.76-2.236-.873-1.009-1.265-2.19-.968-3.301.59-2.2 3.663-3.29 6.863-2.432A8.186 8.186 0 0 1 16.5 5.21M6.42 17.81c.857.99 2.176 1.812 3.761 2.237 3.2.858 6.274-.23 6.863-2.431.233-.868.044-1.779-.465-2.617M3.75 12h16.5"
						/>
					</svg>
				</button>
				<button
					@click="editor.chain().focus().toggleCode().run()"
					:disabled="!editor.can().chain().focus().toggleCode().run()"
					title="code"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('code') }"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-6"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="m6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6a2.25 2.25 0 0 0-2.25-2.25H5.25A2.25 2.25 0 0 0 3 6v12a2.25 2.25 0 0 0 2.25 2.25Z"
						/>
					</svg>
				</button>
				<!--				<button @click="editor.chain().focus().unsetAllMarks().run()">
									Clear marks
								</button>
								<button @click="editor.chain().focus().clearNodes().run()">
									Clear nodes
								</button>
								<button @click="editor.chain().focus().setParagraph().run()"
												:class="{ 'is-active': editor.isActive('paragraph') }">
									Paragraph
								</button>-->

				<button
					@click="editor.chain().focus().toggleBulletList().run()"
					title="bullet list"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('bulletList') }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-width="2"
							d="M9 8h10M9 12h10M9 16h10M4.99 8H5m-.02 4h.01m0 4H5"
						/>
					</svg>
				</button>
				<button
					@click="editor.chain().focus().toggleOrderedList().run()"
					title="ordered list"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('orderedList') }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 6h8m-8 6h8m-8 6h8M4 16a2 2 0 1 1 3.321 1.5L4 20h5M4 5l2-1v6m-2 0h4"
						/>
					</svg>
				</button>
				<button
					@click="editor.chain().focus().toggleCodeBlock().run()"
					title="code block"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('codeBlock') }"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-6"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5"
						/>
					</svg>
				</button>
				<button
					@click="editor.chain().focus().toggleBlockquote().run()"
					title="blockquote"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('blockquote') }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							fill-rule="evenodd"
							d="M6 6a2 2 0 0 0-2 2v3a2 2 0 0 0 2 2h3a3 3 0 0 1-3 3H5a1 1 0 1 0 0 2h1a5 5 0 0 0 5-5V8a2 2 0 0 0-2-2H6Zm9 0a2 2 0 0 0-2 2v3a2 2 0 0 0 2 2h3a3 3 0 0 1-3 3h-1a1 1 0 1 0 0 2h1a5 5 0 0 0 5-5V8a2 2 0 0 0-2-2h-3Z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
				<button title="horizontal line" @click="editor.chain().focus().setHorizontalRule().run()">
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M5 12h14" />
						<path
							stroke="currentColor"
							stroke-linecap="round"
							d="M6 9.5h12m-12-2h12m-12-2h12m-12 13h12m-12-2h12m-12-2h12"
						/>
					</svg>
				</button>
				<!--				<button @click="editor.chain().focus().setHardBreak().run()">
									Hard break
								</button>-->
				<button
					title="undo"
					@click="editor.chain().focus().undo().run()"
					:disabled="!editor.can().chain().focus().undo().run()"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 9h13a5 5 0 0 1 0 10H7M3 9l4-4M3 9l4 4"
						/>
					</svg>
				</button>
				<button
					title="redo"
					@click="editor.chain().focus().redo().run()"
					:disabled="!editor.can().chain().focus().redo().run()"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M21 9H8a5 5 0 0 0 0 10h9m4-10-4-4m4 4-4 4"
						/>
					</svg>
				</button>

				<button
					@click="editor.chain().focus().setColor('var(--M0)').run()"
					title="Main color"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('textStyle', { color: 'var(--M0)' }) }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							fill="currentColor"
							d="M19.9999 18.0661c0 1.6203-1.3431 2.9339-3 2.9339-1.6568 0-3-1.3136-3-2.9339 0-1.6204 3-6.0661 3-6.0661s3 4.4457 3 6.0661Z"
						/>
						<path
							fill="currentColor"
							fill-rule="evenodd"
							d="M10.4817 7.52489 9.12238 10.9817H11.841l-1.3593-3.45681Zm3.7494 4.06961-2.7166-6.90843c-.3694-.93918-1.69627-.93917-2.06558 0L6.76269 11.5173c-.03333.0634-.06004.1309-.07922.2014l-1.28309 3.263h-.41869c-.55229 0-1 .4477-1 1s.44771 1 1 1h2.75c.55228 0 1-.4477 1-1s-.44772-1-1-1h-.18223l.78646-2h4.29158l.3676.9349c.2021.514.7826.7668 1.2966.5647.514-.2021.7668-.7826.5647-1.2966l-.6085-1.5473c-.0053-.0144-.0109-.0287-.0168-.0429Z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>

				<button
					@click="editor.chain().focus().setColor('var(--I0)').run()"
					title="interaktivitäts color"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('textStyle', { color: 'var(--I0)' }) }"
				>
					<svg
						class="w-6 h-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-width="2"
							d="m6.08169 15.9817 1.57292-4m-1.57292 4h-1.1m1.1 0h1.65m-.07708-4 2.72499-6.92967c.0368-.09379.1673-.09379.2042 0l2.725 6.92967m-5.65419 0h-.00607m.00607 0h5.65419m0 0 .6169 1.569m5.1104 4.453c0 1.1025-.8543 1.9963-1.908 1.9963s-1.908-.8938-1.908-1.9963c0-1.1026 1.908-4.1275 1.908-4.1275s1.908 3.0249 1.908 4.1275Z"
						/>
					</svg>
				</button>

				<button
					@click="editor.chain().focus().setColor('var(--A0)').run()"
					title="Akzentfarbe"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('textStyle', { color: 'var(--A0)' }) }"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-6"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M4.098 19.902a3.75 3.75 0 0 0 5.304 0l6.401-6.402M6.75 21A3.75 3.75 0 0 1 3 17.25V4.125C3 3.504 3.504 3 4.125 3h5.25c.621 0 1.125.504 1.125 1.125v4.072M6.75 21a3.75 3.75 0 0 0 3.75-3.75V8.197M6.75 21h13.125c.621 0 1.125-.504 1.125-1.125v-5.25c0-.621-.504-1.125-1.125-1.125h-4.072M10.5 8.197l2.88-2.88c.438-.439 1.15-.439 1.59 0l3.712 3.713c.44.44.44 1.152 0 1.59l-2.879 2.88M6.75 17.25h.008v.008H6.75v-.008Z"
						/>
					</svg>
				</button>
			</div>

			<div class="gap-1 flex flex-wrap">
				<button
					@click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
					title="Titel"
					class="p-1"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('heading', { level: 1 }) }"
				>
					<span class="font-bold text-lg">AaBb</span>
					<hr />
					<span class="text-sm">Titel</span>
				</button>

				<button
					@click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
					title="Überschrift 2"
					class="p-1"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('heading', { level: 2 }) }"
				>
					<span class="font-bold">AaBb</span>
					<hr />
					<span class="text-sm">Überschrift 2</span>
				</button>

				<button
					@click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
					title="Überschrift 3"
					class="p-1"
					:class="{ 'bg-I0 rounded-sm': editor.isActive('heading', { level: 3 }) }"
				>
					<span class="">AaBb</span>
					<hr />
					<span class="text-sm">Überschrift 3</span>
				</button>
			</div>
		</div>

		<editor-content :editor="editor" class="prose-custom" />
	</div>
</template>
