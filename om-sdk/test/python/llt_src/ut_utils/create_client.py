from flask import Flask, Blueprint, g

app = Flask(__name__)


@app.before_request
def before_request():
    # 给 g.user_id 打桩
    g.user_id = 123


class TestUtils:
    user_id = "12345"


def get_client(blueprint: Blueprint):
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()
