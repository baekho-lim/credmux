"""credmux Telegram bot entry point.

Usage:
    pip install -r bot/requirements.txt
    TELEGRAM_BOT_TOKEN=xxx DEMO_HOME=./demo python3 bot/bot.py

Optional env vars:
    CREDMUX_BINARY     path to credmux binary (default: ./credmux)
    DEMO_HOME          forwarded to subprocess for demo mode
    RSS_POLL_INTERVAL  seconds between Vercel RSS polls (default: 60)
"""

from __future__ import annotations

import logging
import os
import sys

from dotenv import load_dotenv
from telegram.ext import Application, CallbackQueryHandler, CommandHandler

from handlers import (
    cmd_help,
    cmd_start,
    cmd_testbreach,
    on_callback,
    rss_poll,
)


def main() -> None:
    load_dotenv()
    logging.basicConfig(
        format="%(asctime)s %(name)s [%(levelname)s] %(message)s",
        level=logging.INFO,
    )
    log = logging.getLogger("credmux-bot")

    token = os.getenv("TELEGRAM_BOT_TOKEN")
    if not token:
        sys.exit("TELEGRAM_BOT_TOKEN is required")

    poll_interval = int(os.getenv("RSS_POLL_INTERVAL", "60"))

    app = Application.builder().token(token).build()
    app.add_handler(CommandHandler("start", cmd_start))
    app.add_handler(CommandHandler("help", cmd_help))
    app.add_handler(CommandHandler("testbreach", cmd_testbreach))
    app.add_handler(CallbackQueryHandler(on_callback))

    if app.job_queue is not None:
        app.job_queue.run_repeating(rss_poll, interval=poll_interval, first=10)
        log.info("SENTINEL RSS poller scheduled every %ds", poll_interval)
    else:
        log.warning(
            "JobQueue unavailable — install python-telegram-bot[job-queue] "
            "to enable RSS polling. /testbreach still works.",
        )

    log.info("credmux bot starting (binary=%s, demo_home=%s)",
             os.getenv("CREDMUX_BINARY", "./credmux"),
             os.getenv("DEMO_HOME", "<unset>"))
    app.run_polling()


if __name__ == "__main__":
    main()
