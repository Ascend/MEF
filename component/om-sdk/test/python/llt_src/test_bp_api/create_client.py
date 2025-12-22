from flask import Flask, Blueprint


class TestUtilsSub:
    filename = 'TomcatR_SATA_FSCJ.txt'


class TestUtils:
    headers = {'X-Real-Ip': '127.0.0.1'}
    files = {'imgfile': TestUtilsSub}
    form = {'size': '8'}


def get_client(blueprint: Blueprint):
    app = Flask(__name__)
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()