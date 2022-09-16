import React, { useState, useEffect } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"

var client = new NexivilClient('http://localhost:8000')

function App() {

  const contentList = []

  const [contents, setContents] = useState([]);

  const GetNexivilContent = () => {
    var contentRequest = new ContentRequest();
    contentRequest.setProjectName("blue");
    var stream = client.listContents(contentRequest,{});

    stream.on('data', function(response) {
      console.log("stream")
      contentList.push(response.getContent())
      console.log(contentList)
      setContents([...contents, contentList])
    });

    console.log(contents)

    console.log("🌈");

    const contentItems = contents.map((content) =>
      <li>{content}</li>
    )

    // useEffect(()=>{
    //   GetNexivilContent()
    // },[]);

    return (
      <div>
        {contentItems}
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
