import React, { useState, useEffect } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"

var client = new NexivilClient('http://localhost:8000')

function App() {

  const [contents, setContents] = useState([ {
    "project_name": "example project1",
    "content": "Veniam dolore tempor eiusmod cupidatat deserunt aliquip quis ad sit in exercitation. Non sint elit nulla reprehenderit voluptate ut amet. Dolor occaecat incididunt cupidatat adipisicing culpa in commodo elit laborum.\r\n"
  },
  {
    "project_name": "example project2",
    "content": "Irure non dolor ipsum tempor id est. Sint duis sunt qui anim cillum cillum et enim nisi. Velit ipsum reprehenderit voluptate fugiat ullamco laborum ullamco nostrud laborum duis aliqua dolor reprehenderit do. Culpa occaecat aliqua duis labore commodo dolore velit nisi qui. Sint Lorem laboris aute anim anim. In ipsum nisi enim ullamco minim velit amet cillum proident proident pariatur in laborum. Tempor laboris consequat Lorem excepteur officia quis pariatur ex.\r\n"
  }]);

  const GetContent = () => {
    var contentRequest = new ContentRequest();
    contentRequest.setProjectName("blue");
    console.log(contentRequest);
    var stream = client.listContents(contentRequest,{});
    console.log(stream);

    stream.on('data', function(response) {
      console.log("stream")
    });

    console.log("🌈");

    const listItem = contents.map((content) => <li>{content.project_name} 👩‍🎨 <br/> {content.content} </li>)

    return (
      <div>
        {listItem}
      </div>
    )

  }

  return (
    <div className="content-cont">
      <div className="content">
        <h1>Nexivil Content</h1>
        <GetContent />
      </div>
    </div>
  );
}

export default App;
