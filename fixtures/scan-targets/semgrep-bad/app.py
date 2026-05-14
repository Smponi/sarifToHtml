import os
import pickle
import subprocess

from flask import Flask, request

app = Flask(__name__)
app.config["SECRET_KEY"] = "not-a-real-flask-secret"


@app.route("/run")
def run():
    command = request.args.get("cmd", "id")
    return subprocess.check_output(command, shell=True).decode("utf-8")


@app.route("/load", methods=["POST"])
def load():
    payload = request.get_data()
    return str(pickle.loads(payload))


@app.route("/calc")
def calc():
    expression = request.args.get("q", "1 + 1")
    return str(eval(expression))


if __name__ == "__main__":
    os.environ["AWS_SECRET_ACCESS_KEY"] = "not-a-real-python-secret"
    app.run(host="0.0.0.0", debug=True)
