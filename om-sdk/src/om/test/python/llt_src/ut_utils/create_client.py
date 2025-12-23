from flask import Flask, Blueprint

app = Flask(__name__)


class TestUtils:
    user_id = "12345"


def get_client(blueprint: Blueprint):
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()
