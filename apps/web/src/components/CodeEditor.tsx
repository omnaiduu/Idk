import Editor, { type OnMount } from "@monaco-editor/react";
import type { editor } from "monaco-editor";
import { useCallback, useRef } from "react";

interface CodeEditorProps {
  path: string;
  value: string;
  onChange: (value: string) => void;
  readOnly?: boolean;
}

const VOID = "#14120b";
const CREAM = "#efece6";
const EMBER = "#ff6a2a";

function defineTheme(monaco: typeof import("monaco-editor")) {
  monaco.editor.defineTheme("tiny-gpu-void", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "comment", foreground: "8a877f", fontStyle: "italic" },
      { token: "keyword", foreground: "ff6a2a" },
      { token: "string", foreground: "c4bfb5" },
      { token: "number", foreground: "efece6" },
      { token: "type", foreground: "ff9a6a" },
    ],
    colors: {
      "editor.background": VOID,
      "editor.foreground": CREAM,
      "editorLineNumber.foreground": "#5c584f",
      "editorLineNumber.activeForeground": CREAM,
      "editor.selectionBackground": "#ff6a2a33",
      "editor.lineHighlightBackground": "#efece608",
      "editorCursor.foreground": EMBER,
      "editorWidget.background": "#1a1812",
      "editorWidget.border": "#efece61a",
      "minimap.background": VOID,
      "scrollbarSlider.background": "#efece615",
      "scrollbarSlider.hoverBackground": "#efece625",
    },
  });
}

export function CodeEditor({ path, value, onChange, readOnly }: CodeEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  const handleMount: OnMount = useCallback((ed, monaco) => {
    editorRef.current = ed;
    defineTheme(monaco);
    monaco.editor.setTheme("tiny-gpu-void");
  }, []);

  const language = path.endsWith(".c") ? "c" : path.endsWith(".v") ? "verilog" : "plaintext";

  return (
    <div className="relative min-h-0 flex-1 overflow-hidden rounded-md border border-cream-border bg-void">
      <div className="absolute left-3 top-2 z-10 font-mono text-[10px] text-cream-muted">{path}</div>
      <Editor
        height="100%"
        language={language}
        value={value}
        onChange={(v) => onChange(v ?? "")}
        onMount={handleMount}
        options={{
          readOnly,
          minimap: { enabled: false },
          fontFamily: "Geist Mono, ui-monospace, monospace",
          fontSize: 13,
          lineHeight: 20,
          padding: { top: 28, bottom: 12 },
          scrollBeyondLastLine: false,
          renderLineHighlight: "line",
          smoothScrolling: true,
          tabSize: 2,
          wordWrap: "off",
          automaticLayout: true,
          scrollbar: {
            verticalScrollbarSize: 8,
            horizontalScrollbarSize: 8,
          },
        }}
        loading={
          <div className="flex h-full items-center justify-center text-sm text-cream-muted">
            Loading editor…
          </div>
        }
      />
    </div>
  );
}
