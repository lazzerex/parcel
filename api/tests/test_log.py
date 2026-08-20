import json
import logging

import log


def test_emits_json_with_context(capsys):
    logger = log.get_logger("test.context")
    logger.info("file queued", extra={"context": {"file_id": "abc", "status": "QUEUED"}})

    payload = json.loads(capsys.readouterr().out.strip())

    assert payload["level"] == "INFO"
    assert payload["message"] == "file queued"
    assert payload["file_id"] == "abc"
    assert payload["status"] == "QUEUED"


def test_includes_exception_details(capsys):
    logger = log.get_logger("test.error")
    try:
        raise ValueError("boom")
    except ValueError:
        logger.exception("processing failed")

    payload = json.loads(capsys.readouterr().out.strip())

    assert payload["level"] == "ERROR"
    assert "ValueError: boom" in payload["error"]


def test_handler_is_not_added_twice():
    logger = log.get_logger("test.idempotent")
    log.get_logger("test.idempotent")

    assert len(logger.handlers) == 1
    assert logger.propagate is False


def test_respects_log_level_env(monkeypatch):
    monkeypatch.setenv("LOG_LEVEL", "WARNING")

    logger = log.get_logger("test.level")

    assert logger.level == logging.WARNING
