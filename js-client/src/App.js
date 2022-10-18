import React, { useState, useEffect, useRef } from 'react';
import './App.css';
import { ContentRequest } from "./contentpb/content_pb"
import { NexivilClient } from "./contentpb/content_grpc_web_pb"
import { List as immuList } from "immutable"
import * as moment from "moment"

var client = new NexivilClient('http://localhost:8000')

function App() {

  const [projectName, setProjectName] = useState("");
  const [content, setContent] = useState(immuList());
  const testRef = useRef({ data: undefined, date: undefined })



  useEffect(() => {
    const getNexivilContent = () => {

      var contentRequest = new ContentRequest();
      // contentRequest.setProjectName(projectName);
      var stream = client.nexivilContent(contentRequest, {});

      stream.on('data', function (response) {
        console.log("🌈");

        // console.log(response.getContent());
        console.log(response.getProjectName());
        console.log(response.getDate())
        console.log(moment.utc(response.getDate()).local().format("YYYY-MM-DD hh:mm:ss"))
        let currentData = 0.01, currentDate
        if (testRef.current.date !== (currentDate = response.getDate()))
          setContent(store => store.push({ data: currentData += (0.01 * Math.random()), date: moment.utc(currentDate).local().format("YYYY-MM-DD hh:mm:ss") }));
        setProjectName(response.getProjectName());

        // response.getDate()
      });

    };
    getNexivilContent()
  }, []);

  return (
    <div className="content-cont">
      <div className="content">
        <h1>Nexivil Content</h1>
      </div>

      <div>
        <h2> project name: {projectName} </h2>
        <h2> content:</h2>
      </div>
      <div style={{ display: "flex", flexDirection: "column", width: "100%" }}>
        {content.valueSeq().map(c => <div>{`${c.data} ${c.date}`}</div>)}
      </div>
    </div>
  );
}

export default App;
