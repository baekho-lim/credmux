"""credmux Telegram bot handlers.

Layout:
  - PLATFORMS / KEYWORDS — config
  - SUBSCRIBERS / SEEN_RSS_ENTRIES — in-memory state (single-process bot)
  - render_*() — message builders
  - run_scan() — subprocess wrapper for `credmux breach-drill <p> --json`
  - cmd_*, on_callback — Telegram handlers
  - rss_poll — JobQueue callback (SENTINEL)
"""

from __future__ import annotations

import json
import logging
import os
import subprocess
from typing import Any

import feedparser
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update
from telegram.constants import ParseMode
from telegram.ext import ContextTypes

logger = logging.getLogger(__name__)

CREDMUX = os.getenv("CREDMUX_BINARY", "./credmux")
DEMO_HOME = os.getenv("DEMO_HOME")

PLATFORMS: dict[str, dict[str, Any]] = {
    "vercel": {
        "name": "Vercel",
        "rss": "https://www.vercel-status.com/history.rss",
        "dashboard": "https://vercel.com/account/tokens",
        "rotate_steps": [
            "1️⃣ vercel.com/account/tokens 접속",
            "2️⃣ 영향받은 토큰 'Revoke' 클릭",
            "3️⃣ 'Create' 눌러 새 토큰 발급",
            "4️⃣ 발급된 토큰을 아래에 붙여넣어 주세요.",
        ],
    },
}

KEYWORDS = ("security", "breach", "incident")

SUBSCRIBERS: set[int] = set()
SEEN_RSS_ENTRIES: set[str] = set()


# ---------------------------------------------------------------------------
# Message builders
# ---------------------------------------------------------------------------

def breach_alert_keyboard(platform: str) -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup([[
        InlineKeyboardButton("✅ 지금 스캔", callback_data=f"scan:{platform}"),
        InlineKeyboardButton("⏰ 나중에", callback_data=f"later:{platform}"),
    ]])


def rotate_keyboard(platform: str) -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup([[
        InlineKeyboardButton("🔑 교체 안내", callback_data=f"rotate:{platform}"),
    ]])


def render_breach_alert(meta: dict[str, Any], summary: str) -> str:
    return (
        f"🚨 *{meta['name']} 보안 경고*\n"
        f"{summary}\n\n"
        "내 프로젝트가 영향받았는지 지금 확인할까요?"
    )


def render_findings(payload: dict[str, Any]) -> str:
    if payload.get("error"):
        return f"⚠️ 스캔 오류: {payload['error']}"
    findings = payload.get("findings") or []
    if not findings:
        return "✅ 영향받은 토큰 없음. 안전합니다."
    lines = [f"⚠️ *{len(findings)}개 토큰* 발견:"]
    for f in findings:
        lines.append(f"  • `{f['file']}:{f['line']}`")
        lines.append(f"    {f['detector']}  `{f['raw_masked']}`")
    impact = payload.get("impact_files") or len({f["file"] for f in findings})
    lines.append(f"\n영향: {impact}개 파일")
    if payload.get("dashboard_url"):
        lines.append(f"교체 페이지: {payload['dashboard_url']}")
    return "\n".join(lines)


def render_rotate_guide(meta: dict[str, Any]) -> str:
    steps = "\n".join(meta["rotate_steps"])
    return (
        f"🔑 *{meta['name']} 토큰 교체 안내*\n\n"
        f"{steps}\n\n"
        "새 토큰을 여기에 붙여넣으세요. 안전하게 macOS Keychain에 저장됩니다."
    )


# ---------------------------------------------------------------------------
# subprocess scan
# ---------------------------------------------------------------------------

def run_scan(platform: str) -> tuple[dict[str, Any] | None, str | None]:
    """Invoke `credmux breach-drill <platform> --json` and parse stdout.

    Returns (payload, error_message). On parse failure both fields surface so
    the bot can decide whether to bubble up a friendly note.
    """
    cmd = [CREDMUX, "breach-drill", platform, "--json"]
    env = os.environ.copy()
    if DEMO_HOME:
        env["DEMO_HOME"] = DEMO_HOME
    try:
        proc = subprocess.run(
            cmd, capture_output=True, text=True, env=env, timeout=60,
        )
    except FileNotFoundError:
        return None, f"credmux 바이너리를 찾을 수 없습니다 ({CREDMUX})."
    except subprocess.TimeoutExpired:
        return None, "스캔이 60초 안에 끝나지 않았습니다."

    stdout = (proc.stdout or "").strip()
    if not stdout:
        return None, f"스캔 출력이 비어 있습니다 (exit={proc.returncode}, stderr={proc.stderr.strip()})"
    try:
        return json.loads(stdout), None
    except json.JSONDecodeError as e:
        return None, f"JSON 파싱 실패: {e}"


# ---------------------------------------------------------------------------
# Command handlers
# ---------------------------------------------------------------------------

async def cmd_start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    chat_id = update.effective_chat.id
    SUBSCRIBERS.add(chat_id)
    await update.message.reply_text(
        "🔐 *credmux bot*에 오신 것을 환영합니다.\n\n"
        "외부 플랫폼 보안 사고가 발생하면 즉시 알려드리고, "
        "내 맥에 영향받은 토큰이 있는지 한 번에 확인합니다.\n\n"
        "*명령어*\n"
        "  /testbreach — Vercel 사고 데모 알림\n"
        "  /help — 도움말\n\n"
        "이제부터 자동으로 보안 사고 알림을 받게 됩니다.",
        parse_mode=ParseMode.MARKDOWN,
    )


async def cmd_help(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    await update.message.reply_text(
        "/start — 알림 구독\n"
        "/testbreach — Vercel 데모 알림 (해커톤용)\n"
        "/help — 이 메시지",
    )


async def cmd_testbreach(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    chat_id = update.effective_chat.id
    SUBSCRIBERS.add(chat_id)
    platform = "vercel"
    meta = PLATFORMS[platform]
    summary = "(데모 시뮬) Vercel Status — Security Incident 2026-04-19"
    await context.bot.send_message(
        chat_id=chat_id,
        text=render_breach_alert(meta, summary),
        parse_mode=ParseMode.MARKDOWN,
        reply_markup=breach_alert_keyboard(platform),
    )


async def on_callback(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    query = update.callback_query
    await query.answer()
    data = query.data or ""
    parts = data.split(":", 1)
    if len(parts) != 2:
        return
    action, platform = parts
    meta = PLATFORMS.get(platform)
    if meta is None:
        await query.edit_message_text(f"알 수 없는 플랫폼: {platform}")
        return

    if action == "scan":
        await query.edit_message_text(f"🔎 {meta['name']} 토큰 스캔 중…")
        payload, err = run_scan(platform)
        if err:
            await context.bot.send_message(
                chat_id=query.message.chat_id,
                text=f"⚠️ {err}",
            )
            return
        text = render_findings(payload or {})
        if payload and payload.get("findings"):
            await context.bot.send_message(
                chat_id=query.message.chat_id,
                text=text,
                parse_mode=ParseMode.MARKDOWN,
                reply_markup=rotate_keyboard(platform),
            )
        else:
            await context.bot.send_message(
                chat_id=query.message.chat_id,
                text=text,
            )
        return

    if action == "rotate":
        await context.bot.send_message(
            chat_id=query.message.chat_id,
            text=render_rotate_guide(meta),
            parse_mode=ParseMode.MARKDOWN,
        )
        return

    if action == "later":
        await query.edit_message_text("⏰ 알겠습니다. 필요할 때 다시 알려드릴게요.")
        return

    await query.edit_message_text(f"알 수 없는 액션: {action}")


# ---------------------------------------------------------------------------
# SENTINEL: RSS poll
# ---------------------------------------------------------------------------

async def rss_poll(context: ContextTypes.DEFAULT_TYPE) -> None:
    if not SUBSCRIBERS:
        return
    for platform, meta in PLATFORMS.items():
        url = meta.get("rss")
        if not url:
            continue
        try:
            feed = feedparser.parse(url)
        except Exception as e:  # network hiccup, malformed feed, etc.
            logger.warning("RSS fetch failed for %s: %s", platform, e)
            continue
        for entry in feed.entries[:5]:
            entry_id = f"{platform}:{getattr(entry, 'id', None) or getattr(entry, 'link', '')}"
            if entry_id in SEEN_RSS_ENTRIES:
                continue
            haystack = f"{getattr(entry, 'title', '')} {getattr(entry, 'summary', '')}".lower()
            SEEN_RSS_ENTRIES.add(entry_id)
            if not any(k in haystack for k in KEYWORDS):
                continue
            summary = (
                f"{getattr(entry, 'title', 'Incident detected')}"
                f" — {getattr(entry, 'published', '')}"
            )
            for chat_id in list(SUBSCRIBERS):
                try:
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text=render_breach_alert(meta, summary),
                        parse_mode=ParseMode.MARKDOWN,
                        reply_markup=breach_alert_keyboard(platform),
                    )
                except Exception as e:
                    logger.warning("send_message failed (chat=%s): %s", chat_id, e)
