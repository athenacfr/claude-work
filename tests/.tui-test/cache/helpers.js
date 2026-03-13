//# hash=5bbb1055a4a471cdd88791a5ae71e587
//# sourceMappingURL=helpers.js.map

function asyncGeneratorStep(gen, resolve, reject, _next, _throw, key, arg) {
    try {
        var info = gen[key](arg);
        var value = info.value;
    } catch (error) {
        reject(error);
        return;
    }
    if (info.done) {
        resolve(value);
    } else {
        Promise.resolve(value).then(_next, _throw);
    }
}
function _async_to_generator(fn) {
    return function() {
        var self = this, args = arguments;
        return new Promise(function(resolve, reject) {
            var gen = fn.apply(self, args);
            function _next(value) {
                asyncGeneratorStep(gen, resolve, reject, _next, _throw, "next", value);
            }
            function _throw(err) {
                asyncGeneratorStep(gen, resolve, reject, _next, _throw, "throw", err);
            }
            _next(undefined);
        });
    };
}
function _ts_generator(thisArg, body) {
    var f, y, t, _ = {
        label: 0,
        sent: function() {
            if (t[0] & 1) throw t[1];
            return t[1];
        },
        trys: [],
        ops: []
    }, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype), d = Object.defineProperty;
    return d(g, "next", {
        value: verb(0)
    }), d(g, "throw", {
        value: verb(1)
    }), d(g, "return", {
        value: verb(2)
    }), typeof Symbol === "function" && d(g, Symbol.iterator, {
        value: function() {
            return this;
        }
    }), g;
    function verb(n) {
        return function(v) {
            return step([
                n,
                v
            ]);
        };
    }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while(g && (g = 0, op[0] && (_ = 0)), _)try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [
                op[0] & 2,
                t.value
            ];
            switch(op[0]){
                case 0:
                case 1:
                    t = op;
                    break;
                case 4:
                    _.label++;
                    return {
                        value: op[1],
                        done: false
                    };
                case 5:
                    _.label++;
                    y = op[1];
                    op = [
                        0
                    ];
                    continue;
                case 7:
                    op = _.ops.pop();
                    _.trys.pop();
                    continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) {
                        _ = 0;
                        continue;
                    }
                    if (op[0] === 3 && (!t || op[1] > t[0] && op[1] < t[3])) {
                        _.label = op[1];
                        break;
                    }
                    if (op[0] === 6 && _.label < t[1]) {
                        _.label = t[1];
                        t = op;
                        break;
                    }
                    if (t && _.label < t[2]) {
                        _.label = t[2];
                        _.ops.push(op);
                        break;
                    }
                    if (t[2]) _.ops.pop();
                    _.trys.pop();
                    continue;
            }
            op = body.call(thisArg, _);
        } catch (e) {
            op = [
                6,
                e
            ];
            y = 0;
        } finally{
            f = t = 0;
        }
        if (op[0] & 5) throw op[1];
        return {
            value: op[0] ? op[1] : void 0,
            done: true
        };
    }
}
var _process_env_CW_TEST_ROOT;
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";
import { test, expect } from "@microsoft/tui-test";
// tui-test compiles to a cache dir, so we resolve the binary path
// relative to the project root using an env var or a known absolute path.
var PROJECT_ROOT = (_process_env_CW_TEST_ROOT = process.env.CW_TEST_ROOT) !== null && _process_env_CW_TEST_ROOT !== void 0 ? _process_env_CW_TEST_ROOT : join(dirname(fileURLToPath(import.meta.url)), "..");
// Resolved path to the cw binary (built by `make build` before tests run)
export var CW_BIN = join(PROJECT_ROOT, "bin", "cw");
// Create an isolated temp environment for a test suite.
// Returns paths and a cleanup function.
export function createTestEnv() {
    var root = mkdtempSync(join(tmpdir(), "cw-test-"));
    var projectsDir = join(root, "projects");
    var dataDir = join(root, "data");
    mkdirSync(projectsDir, {
        recursive: true
    });
    mkdirSync(dataDir, {
        recursive: true
    });
    return {
        root: root,
        projectsDir: projectsDir,
        dataDir: dataDir,
        env: {
            CW_PROJECTS_DIR: projectsDir,
            CW_DATA_DIR: dataDir,
            // Prevent cw from picking up real user config
            HOME: root
        },
        cleanup: function cleanup() {
            rmSync(root, {
                recursive: true,
                force: true
            });
        }
    };
}
// Create a fake project inside the test env's projects dir.
// Initializes a bare git repo so cw recognizes it.
export function createFakeProject(projectsDir, name, opts) {
    var _ref;
    var projectDir = join(projectsDir, name);
    mkdirSync(projectDir, {
        recursive: true
    });
    // Create .cw metadata dir
    var cwDir = join(projectDir, ".cw");
    mkdirSync(cwDir, {
        recursive: true
    });
    if (opts === null || opts === void 0 ? void 0 : opts.metadata) {
        writeFileSync(join(cwDir, "metadata.json"), JSON.stringify(opts.metadata, null, 2));
    }
    // Create repo subdirectories with git init
    var repos = (_ref = opts === null || opts === void 0 ? void 0 : opts.repos) !== null && _ref !== void 0 ? _ref : [
        "repo-one"
    ];
    var _iteratorNormalCompletion = true, _didIteratorError = false, _iteratorError = undefined;
    try {
        for(var _iterator = repos[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true){
            var repo = _step.value;
            var repoDir = join(projectDir, repo);
            mkdirSync(repoDir, {
                recursive: true
            });
            execSync("git init -q && git commit --allow-empty -m init -q", {
                cwd: repoDir,
                stdio: "ignore"
            });
        }
    } catch (err) {
        _didIteratorError = true;
        _iteratorError = err;
    } finally{
        try {
            if (!_iteratorNormalCompletion && _iterator.return != null) {
                _iterator.return();
            }
        } finally{
            if (_didIteratorError) {
                throw _iteratorError;
            }
        }
    }
    return projectDir;
}
// Configure tui-test to launch the cw binary with isolated env.
// Call this at the top of each test file.
export function useCw(env) {
    test.use({
        program: {
            file: CW_BIN
        },
        rows: 24,
        columns: 80,
        env: env
    });
}
// Wait for the TUI to be fully rendered by checking for the footer bar.
// The footer (permissions/compact) is the last thing rendered, so its presence
// indicates the full TUI frame has been drawn.
export function waitForReady(terminal) {
    return _async_to_generator(function() {
        return _ts_generator(this, function(_state) {
            switch(_state.label){
                case 0:
                    return [
                        4,
                        expect(terminal.getByText(/permissions/g)).toBeVisible()
                    ];
                case 1:
                    _state.sent();
                    return [
                        2
                    ];
            }
        });
    })();
}
