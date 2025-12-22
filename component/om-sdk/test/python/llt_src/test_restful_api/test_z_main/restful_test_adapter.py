import os

from unittest.mock import patch
from flask.testing import FlaskClient

from common.ResourceDefV1.service_root import RfServiceRoot
from ut_utils.models import MockPrivilegeAuth
from ibma_redfish_urls import RedfishURIs


class RestfulTestAdapter:
    url_prefix = ""

    def __init__(self):
        self.token_patch = None
        self.flask_patch = None

    @staticmethod
    def get_flask_test_client() -> FlaskClient:
        flask_app = RedfishURIs.APP
        flask_client = flask_app.test_client()
        return flask_client

    def get_url_prefix(self):
        return self.url_prefix

    def setup_class(self):
        self.token_patch = patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth)
        self.token_patch.start()
        self.flask_patch = patch("ibma_redfish_urls.RedfishURIs._start_flask", return_value=None)
        self.flask_patch.start()

        root_dir = os.path.dirname(
            os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))))
        project_dir = os.path.join(root_dir, "install/software/RedfishServer")
        profile_mockup_path = os.path.join(project_dir, "common", "MockupData", "iBMAServerV1")
        root_path = os.path.normpath("redfish/v1")
        root = RfServiceRoot(profile_mockup_path, root_path)
        RedfishURIs.rf_api_ibma_server(root, None, None, None)

    def teardown_class(self):
        if self.token_patch:
            self.token_patch.stop()

        if self.flask_patch:
            self.flask_patch.stop()
