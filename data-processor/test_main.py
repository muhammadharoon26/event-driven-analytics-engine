import unittest
from datetime import datetime
from pydantic import ValidationError
from main import EventModel


class TestEventModel(unittest.TestCase):
    def test_validation_success(self):
        payload = {
            "user_id": "user-123",
            "action": "scan_started",
            "timestamp": "2026-03-09T21:19:19Z",
        }
        event = EventModel(**payload)
        self.assertEqual(event.user_id, "user-123")
        self.assertEqual(event.action, "scan_started")

    def test_validation_failure_missing_field(self):
        payload = {
            "user_id": "user-123",
            # missing action
            "timestamp": "2026-03-09T21:19:19Z",
        }
        with self.assertRaises(ValidationError):
            EventModel(**payload)


if __name__ == "__main__":
    unittest.main()
