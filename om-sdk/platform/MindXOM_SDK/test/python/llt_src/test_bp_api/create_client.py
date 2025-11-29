from flask import Flask, Blueprint


def get_client(blueprint: Blueprint):
    app = Flask(__name__)
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()