import React, { useState, useEffect } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"

var client = new NexivilClient('http://localhost:8000')

function App() {

  const [projectName, setProjectName] = useState("");
  const [content, setContent] = useState("");

  const getNexivilContent = () => {
  
    var contentRequest = new ContentRequest();
    // contentRequest.setProjectName(projectName);
    var stream = client.nexivilContent(contentRequest,{});
  
    stream.on('data', function(response) {
      console.log("🌈");

      console.log(response.getContent());
      console.log(response.getProjectName());
  
      setContent(response.getContent());
      setProjectName(response.getProjectName());
    });

  };

  useEffect(()=>{
    getNexivilContent()
  },[]);

  return (
    <div className="content-cont">
      <div className="content">
        <h1>Nexivil Content</h1>
      </div>

      <div>
        <h2> project name: {projectName} </h2>
        <h2> content: {content} </h2>
      </div>
      
    </div>
  );
}

export default App;
