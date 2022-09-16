import React, { useState, useEffect } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"

var client = new NexivilClient('http://localhost:8000')

function App() {

  const [projectName, setProjectName] = useState([])
  const [contents, setContents] = useState([]);

  const GetNexivilContent = () => {
    var contentRequest = new ContentRequest();
    contentRequest.setProjectName("blue");
    var stream = client.listContents(contentRequest,{});

    console.log(stream);

    stream.on('data', function(response) {
      console.log("stream")
      console.log(response.getContent())
      setProjectName([...projectName, response.getProjectName()])
      setContents([...contents, response.getContent()])
    });

    console.log("🌈");

    // const listItem = contents.map((content) => <li key={content.id}>{content.project_name} 👩‍🎨 <br/> {content.content} </li>)

    return (
      <div>
        {projectName}<br/>
        {contents}
      </div>
    )

  }

  return (
    <div className="content-cont">
      <div className="content">
        <h1>Nexivil Content</h1>
        <GetNexivilContent />
      </div>
    </div>
  );
}

export default App;
