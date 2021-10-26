from os.path import dirname, join, realpath
import re
from yaml import safe_load

ansible_dir = join(dirname(realpath(__file__)), "..", "..", "ansible")

with open(join(ansible_dir, "requirements.yml")) as fp:
    reqs = safe_load(fp.read())

with open(join(ansible_dir, "collections", "requirements.yml")) as fp:
    offline_reqs = safe_load(fp.read())

offline_collections = dict(
    sorted(
        [
            (
                re.sub(r"-(\d+\.){3}tar.gz$", "", coll["name"]).replace("-", "."),
                coll["version"],
            )
            for coll in offline_reqs["collections"]
        ],
        key=lambda coll: coll[0],
    )
)

collections = dict(
    sorted(
        [(coll["name"], coll["version"]) for coll in reqs["collections"]],
        key=lambda coll: coll[0],
    )
)

if collections != offline_collections:
    print(
        "Vendored collections do not match actual, please run `ansible-galaxy collection download -r ansible/requirements.yml -p ansible/collections/`."
    )
    print(f"Expected: {collections}")
    print(f"Vendored: {offline_collections}")
    exit(1)
