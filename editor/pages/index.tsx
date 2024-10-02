import React from 'react';
import dynamic from 'next/dynamic';

const CodeEditor = dynamic(() => import('../components/CodeEditor'), { ssr: false });

const Home: React.FC = () => {
  return (
    <div className="text-center">
      <h1 className="text-3xl font-bold my-4">Go Code Editor</h1>
      <CodeEditor />
    </div>
  );
};

export default Home;
