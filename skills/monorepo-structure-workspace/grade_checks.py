#!/usr/bin/env python3
"""Programmatic assertion checks for monorepo-structure evals 1, 2, 4.

Usage: grade_checks.py <eval_dir> <run_subdir> [spawn_epoch]
  eval_dir: e.g. .../iteration-1/eval-1-greenfield-scaffold
  run_subdir: with_skill | without_skill
  spawn_epoch: optional; if given, eval-2 fixture files must have mtime < spawn_epoch

Prints JSON: {"assertions": [{"text": ..., "passed": bool, "evidence": str}]}
"""
import json
import os
import re
import sys
from pathlib import Path

FIXTURE = Path("/home/blkst8/src/github/blkst8/skills/skills/monorepo-structure-workspace/fixtures/legacy-app")


def outdir(eval_dir, run_subdir):
    p = Path(eval_dir) / run_subdir / "outputs"
    # runs may nest the project one level down (outputs/project/...)
    if not any(p.rglob("package.json")) and not any(p.rglob("go.mod")):
        return p
    return p


def find_one(root: Path, pattern: str):
    """First path in root matching glob pattern, or None."""
    matches = sorted(root.rglob(pattern))
    return matches[0] if matches else None


def find_proj(root: Path):
    """Locate the actual project root: dir containing pnpm-workspace.yaml,
    turbo.json, or package.json with workspaces, possibly nested."""
    for candidate in [root] + [d for d in root.rglob("*") if d.is_dir()]:
        if (candidate / "pnpm-workspace.yaml").exists() or (candidate / "turbo.json").exists():
            return candidate
    pkg = find_one(root, "package.json")
    return pkg.parent if pkg else root


def read(path):
    try:
        return Path(path).read_text(errors="replace")
    except Exception:
        return ""


def has_dir(parent: Path, name):
    for d in parent.rglob(name):
        if d.is_dir():
            return d
    return None


def grep(root: Path, pattern, glob="**/*"):
    rx = re.compile(pattern, re.IGNORECASE)
    for f in root.rglob("*"):
        if f.is_file() and f.suffix in ("", ".md", ".txt", ".log"):
            continue
        try:
            if rx.search(f.read_text(errors="replace")):
                return f
        except Exception:
            continue
    return None


def check(text, passed, evidence):
    return {"text": text, "passed": bool(passed), "evidence": str(evidence)[:300]}


def grade_eval1(root):
    proj = find_proj(root)
    res = []

    apps = has_dir(proj, "apps")
    web = find_one(proj, "apps*/web*") or has_dir(proj, "web")
    api = find_one(proj, "apps*/api*") or has_dir(proj, "api")
    pkgs = has_dir(proj, "packages")
    res.append(check(
        "apps/ contains web and api; packages/ exists",
        (apps or (web and api)) and pkgs,
        f"apps={apps} web={web} api={api} packages={pkgs}"))

    ws = read(proj / "pnpm-workspace.yaml")
    ok = "apps/*" in ws.replace('"', '') and "packages/*" in ws.replace('"', '')
    res.append(check("pnpm-workspace.yaml globs apps/* and packages/*", ok, ws.strip()[:120]))

    turbo = read(proj / "turbo.json")
    ok = "turbo.json" in str(list(proj.glob('turbo.json'))) and "^build" in turbo
    res.append(check("turbo.json build task with dependsOn ^build", ok,
                     turbo.replace("\n", " ")[:120]))

    types_pkg = find_one(proj, "packages*/types*/package.json") or find_one(proj, "packages*/types*/src*")
    deep = grep(proj, r'\.\./\.\./\.\./.*packages')
    res.append(check("shared types package exists; no deep relative imports into packages",
                     types_pkg is not None and deep is None,
                     f"types_pkg={types_pkg} deep_import={deep}"))

    web_env = list((proj).rglob("apps/*web*/.env.example")) + list(proj.rglob("apps/*web*/.env*.example*"))
    api_env = list((proj).rglob("apps/*api*/.env.example")) + list(proj.rglob("apps/*api*/.env*.example*"))
    root_env = proj / ".env"
    mixed = root_env.exists() and "DATABASE" in read(root_env) and ("API_URL" in read(root_env))
    res.append(check("per-app .env.example for web and api; no mixed root .env",
                     web_env and api_env and not mixed,
                     f"web_env={web_env[:1]} api_env={api_env[:1]} root_env={root_env.exists()}"))

    pkg = json.loads(read(proj / "package.json") or "{}")
    scripts = pkg.get("scripts", {})
    ok = any("turbo" in str(s) for s in scripts.values()) and all(
        k in scripts for k in ("dev", "build", "test"))
    res.append(check("root package.json scripts delegate to turbo for dev/build/test", ok,
                     json.dumps(scripts)[:150]))

    dc = read(proj / "docker-compose.yml") or read(proj / "docker-compose.yaml") or read(proj / "compose.yml")
    ok = "postgres" in dc.lower() and ("5432" in dc or "image" in dc)
    res.append(check("docker-compose at root provides Postgres", ok, dc.strip()[:120]))
    return res


def grade_eval2(root, spawn_epoch=None):
    proj = find_proj(root)
    res = []

    blob = "\n".join(f.read_text(errors="replace") for f in proj.rglob("*")
                     if f.is_file() and f.suffix in (".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"))
    symbols = {
        "hashPassword": r"hashPassword",
        "verifyPassword": r"verifyPassword",
        "scrypt hashing": r"scryptSync|scrypt",
        "timing-safe compare": r"timingSafeEqual",
        "UserModel/list": r"list\s*\(|SELECT id, email, name",
        "UserModel.create": r"INSERT INTO users|UserModel\.create",
        "login route": r"login|invalid credentials",
        "email-taken check": r"email taken|409",
        "fetchUsers": r"fetchUsers",
        "signup": r"signup|sign ?up",
        "formatDate": r"formatDate",
        "relativeTime": r"relativeTime",
    }
    missing = [k for k, rx in symbols.items() if not re.search(rx, blob)]
    res.append(check("no logic lost: all original functions/routes survive",
                     not missing, f"missing={missing}" if missing else "all 12 symbol groups present"))

    dto_defs = re.findall(r"(?:interface|type)\s+UserDto", blob)
    in_pkg = any("packages" in str(f) for f in proj.rglob("*")
                 if f.is_file() and re.search(r"UserDto", read(f)))
    res.append(check("UserDto defined once in a packages/ dir and referenced by apps",
                     len(dto_defs) <= 1 and in_pkg,
                     f"defs={len(dto_defs)} in_packages={in_pkg}"))

    web_env = list(proj.rglob("apps/*web*/.env*")) or list(proj.rglob("apps/*web*/.env.example*"))
    api_env = list(proj.rglob("apps/*api*/.env*"))
    root_env = proj / ".env"
    mixed = root_env.exists() and "DATABASE" in read(root_env) and "API_URL" in read(root_env)
    res.append(check("root .env split into per-app env files; no mixed root .env",
                     web_env and api_env and not mixed,
                     f"web_env={[str(p) for p in web_env[:2]]} api_env={[str(p) for p in api_env[:2]]} root_mixed={mixed}"))

    pkgs = has_dir(proj, "packages")
    db_leak = None
    if pkgs:
        for f in pkgs.rglob("*"):
            if f.is_file():
                t = read(f)
                if re.search(r"new Pool|INSERT INTO|from \"pg\"|require\(.pg.\)", t):
                    db_leak = f
                    break
    res.append(check("database code (pg/Pool/SQL) stays inside apps/api, not in packages/",
                     db_leak is None, f"leak={db_leak}"))

    ws = (proj / "pnpm-workspace.yaml").exists()
    tj = (proj / "turbo.json").exists()
    rp = (proj / "package.json").exists()
    res.append(check("pnpm-workspace.yaml + turbo.json + root package.json present",
                     ws and tj and rp, f"ws={ws} turbo={tj} root_pkg={rp}"))

    untouched = True
    detail = "fixture missing or spawn_epoch not provided"
    if FIXTURE.exists():
        problems = []
        for f in FIXTURE.rglob("*"):
            if f.is_file():
                if spawn_epoch and f.stat().st_mtime > float(spawn_epoch):
                    problems.append(f"{f.name} mtime={f.stat().st_mtime}")
        if not list(FIXTURE.rglob("*")):
            problems.append("fixture empty")
        untouched = not problems
        detail = f"problems={problems}" if problems else f"{len(list(FIXTURE.rglob('*')))} entries, none modified"
    res.append(check("original fixture untouched", untouched, detail))
    return res


def grade_eval4(root, spawn_epoch=None):
    proj = find_proj(root)
    res = []

    gomod = find_one(proj, "apps*/api*/go.mod") or find_one(proj, "go.mod")
    api_dir = gomod.parent if gomod else None
    res.append(check("apps/api is a native Go module (go.mod)", gomod is not None and api_dir is not None,
                     f"go.mod={gomod}"))

    ok = True
    detail = "no package.json in api"
    if api_dir:
        pkgj = api_dir / "package.json"
        if pkgj.exists():
            p = json.loads(read(pkgj) or "{}")
            deps = p.get("dependencies") or {}
            dev = p.get("devDependencies") or {}
            ok = len(deps) == 0 and len(dev) == 0
            detail = f"deps={deps} dev={dev}"
    res.append(check("no JS dependencies injected into the Go app", ok, detail))

    sharing = grep(proj, r"openapi|proto(buf)?|json schema|codegen|generat(e|ed) (types|client)|oapi")
    res.append(check("cross-language sharing via schemas/codegen (openapi/proto/generate)",
                     sharing is not None, f"match={sharing}"))

    mk = find_one(proj, "Makefile") or find_one(proj, "makefile") or find_one(proj, "justfile")
    wrapper = None
    if api_dir and (api_dir / "package.json").exists():
        wrapper = api_dir / "package.json"
    res.append(check("one orchestration story: Makefile/justfile or thin package.json wrapper",
                     mk is not None or wrapper is not None, f"makefile={mk} wrapper={wrapper}"))

    apps = has_dir(proj, "apps")
    pkgs = has_dir(proj, "packages")
    scraper = grep(proj, r"scraper") or has_dir(proj, "scraper")
    res.append(check("apps/ + packages/ separation; planned home for python scraper",
                     (apps or pkgs) is not None and scraper is not None,
                     f"apps={apps} packages={pkgs} scraper={scraper}"))

    web_env = list(proj.rglob("apps/*web*/.env*"))
    api_env = list(proj.rglob("apps/*api*/.env*")) or list(proj.rglob("apps/*api*/*.env*"))
    dc = read(proj / "docker-compose.yml") or read(proj / "docker-compose.yaml") or read(proj / "compose.yml")
    ok = (web_env or api_env) and "postgres" in dc.lower()
    res.append(check("per-app env files + docker-compose Postgres at root", ok,
                     f"web_env={bool(web_env)} api_env={bool(api_env)} compose={'postgres' in dc.lower()}"))
    return res


def main():
    eval_dir, run_subdir = sys.argv[1], sys.argv[2]
    spawn_epoch = sys.argv[3] if len(sys.argv) > 3 else None
    root = Path(eval_dir) / run_subdir / "outputs"
    eval_name = Path(eval_dir).name
    if "eval-1" in eval_name:
        results = grade_eval1(root)
    elif "eval-2" in eval_name:
        results = grade_eval2(root, spawn_epoch)
    elif "eval-4" in eval_name:
        results = grade_eval4(root)
    else:
        results = []
    print(json.dumps({"eval": eval_name, "run": run_subdir, "assertions": results}, indent=2))


if __name__ == "__main__":
    main()
