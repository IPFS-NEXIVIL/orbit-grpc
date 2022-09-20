import React, { useState, useEffect } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"

var client = new NexivilClient('http://localhost:8000')

function ChangeProject(props) {
  return (
    <div>
      <h3>🚀 Change Project</h3>
      <form onSubmit={event=>{
        event.preventDefault();
        const project = event.target.project.value;
        console.log(project);
        props.onChange(project);
      }}>
        <p><input type="text" name="project" placeholder="project name"/></p>
        <p><input type="submit" value="🗳"></input></p>
      </form>
    </div>
  )
}

function App() {

  const [projectName, setProjectName] = useState("");
  const [contents, setContents] = useState([]);

  const GetNexivilContent = () => {
  
    var contentRequest = new ContentRequest();
    contentRequest.setProjectName(projectName);
    var stream = client.listContents(contentRequest,{});
  
    stream.on('data', function(response) {
      console.log(response);
      // const contentList = [];
      // contentList.push(response.getContent());
  
      setContents(c=>[...c, response.getContent()]);
    });
  
    console.log("🌈");

    return () => {
    };

  };

  useEffect(()=>{
    GetNexivilContent()
  },[]);

  return (
    <div className="content-cont">
      <div className="content">
        <h1>Nexivil Content</h1>
        <ChangeProject onChange={(project)=>{
          const newProject = project;
          setProjectName(newProject);
        }}></ChangeProject>
      </div>

      <div>
        {contents.map((content, idx) =>
          <div key={idx}>
            <span>{content}</span>
          </div>
        )}
      </div>
      
    </div>
  );
}

export default App;
