import React, { useState } from 'react';
import Editor from '@monaco-editor/react';
import axios from 'axios';

const CodeEditor: React.FC = () => {
  const [code, setCode] = useState("// Write your Go code here\n");
  const [output, setOutput] = useState("");

  const handleEditorChange = (value: string | undefined) => {
    setCode(value || "");
  };

  const runCode = async () => {
    try {
      const response = await axios.post("http://localhost:8080/execute", { code });
      setOutput(response.data.output);
    } catch (error) {
      setOutput("Error executing code");
    }
  };

  return (
    <div className="flex flex-col w-full h-screen">
      <Editor
        height="70vh"
        defaultLanguage="go"
        value={code}
        onChange={handleEditorChange}
        theme="vs-dark"
      />
      <button onClick={runCode} className="mt-2 p-2 bg-blue-500 text-white rounded">
        Run Code
      </button>
      <div className="mt-4 p-4 bg-gray-100">
        <h3>Output:</h3>
        <pre>{output}</pre>
      </div>
    </div>
  );
};

export default CodeEditor;
