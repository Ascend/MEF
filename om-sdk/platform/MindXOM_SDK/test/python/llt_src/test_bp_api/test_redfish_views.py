from unittest.mock import patch
from flask.testing import FlaskClient


from test_bp_api.create_client import get_client
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from redfish_service.redfish_views import public_bp, SchemaCollction

client: FlaskClient = get_client(public_bp)


def test_rf_version_service():
    response = client.get("/redfish")
    assert response.status_code == 200
    assert "v1" in response.get_json(force=True)


def test_rf_service_collection():
    response = client.get("/redfish/v1")
    assert response.status_code == 200
    assert "#ServiceRoot.v1_9_0.ServiceRoot" in response.get_json(force=True).values()


def test_rf_metadata_service():
    response = client.get("/redfish/v1/$metadata")
    assert response.status_code == 200


def test_rf_odata_service():
    response = client.get("/redfish/v1/odata")
    assert response.status_code == 200
    assert "/redfish/v1/$metadata" in response.get_json(force=True).values()


def test_rf_schema_collection():
    response = client.get("/redfish/v1/JSONSchemas")
    schema_collection = [member.value for member in SchemaCollction]
    assert response.status_code == 200
    assert len(schema_collection) == response.get_json(force=True)["Members@odata.count"]


def test_rf_schema_info_success():
    schema = "MessageRegistry.v1_0_0"
    response = client.get(f"/redfish/v1/JSONSchemas/{schema}")
    assert response.status_code == 200
    assert schema in response.get_json(force=True).values()


def test_rf_schema_info_failed():
    schema = "MessageRegistry.v1_0_1"
    response = client.get(f"/redfish/v1/JSONSchemas/{schema}")
    assert response.status_code == 400
    assert "error" in response.get_json(force=True)


