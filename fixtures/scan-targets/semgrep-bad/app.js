const express = require("express");
const childProcess = require("child_process");

const app = express();
const apiKey = "not-a-real-js-token-for-sarif-fixtures";

app.get("/run", (req, res) => {
  const command = req.query.cmd || "id";
  childProcess.exec(command, (error, stdout) => {
    res.send(error ? String(error) : stdout);
  });
});

app.get("/eval", (req, res) => {
  res.send(String(eval(req.query.value || "1 + 1")));
});

app.listen(3000, () => {
  console.log(`fixture started with ${apiKey}`);
});
