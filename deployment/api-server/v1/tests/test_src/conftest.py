# pylint: disable=missing-docstring
import logging
import codecs
import json
import copy

import pytest
import requests
import sys
import os

sys.path.append(os.path.join(os.path.dirname(__file__), '../../../'))
import service_conf
from requests.packages.urllib3.exceptions import (InsecureRequestWarning,
                                                  InsecurePlatformWarning)
from sqlalchemy import create_engine
from paramiko.client import SSHClient, AutoAddPolicy

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)
requests.packages.urllib3.disable_warnings(InsecurePlatformWarning)

# pylint: disable=invalid-name
root = logging.getLogger()
if root.handlers:
    for handler in root.handlers:
        root.removeHandler(handler)
logging.basicConfig(level=logging.DEBUG,
                    format='[%(asctime)s][%(threadName)10.1s][%(levelname).1s]'
                    '[%(name)s.%(funcName)s:%(lineno)s] : %(message)s')


modules = []
copy_modules = []
#case 117 temporarily is disabled for test
exclude_dict = {
    'vip': [0, 1, 7, 8, 9, 11, 15, 19, 20, 21, 24, 25, 26, 27, 28, 29, 30, 31, 34, 40, 42, 43, 54, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 87,
    91, 92, 94, 95, 97, 100, 101, 102, 104, 107, 108, 109, 110, 111, 112, 115, 116, 117, 118, 119, 120,
    121, 123, 124, 125]
}


# pylint: disable=missing-docstring
def pytest_addoption(parser):
    parser.addoption('--host', action='store', default='dev1.emotibot.com')
    parser.addoption('--solr_host', action='store')
    parser.addoption('--db_url', action='store',
                     default='logstash-sh.emotibot.com',
                     help='chat log database url')
    parser.addoption('--db_user', action='store', default='root')
    parser.addoption('--db_pass', action='store', default='password')
    parser.addoption('--ssh_user', action='store', default='deployer')
    parser.addoption('--ssh_pass', action='store', default='Emotibot1')
    parser.addoption('--test_case', action='store', default='./test_case.json')
    parser.addoption('--env', action='store', default='dev')


# pylint: disable=missing-docstring
def pytest_generate_tests(metafunc):
    if 'test_bft' in metafunc.function.__name__:
        # pylint: disable=invalid-name
        with codecs.open(metafunc.config.option.test_case, 'r', 'utf-8') as f:
            data = json.load(f)
            metafunc.parametrize("case", data)

    # elif 'test_cmbc' in metafunc.function.__name__:
    #     with codecs.open('./test_case_cmbc_faq_small.json', 'r', 'utf-8') as f:
    #         data = json.load(f)
    #         metafunc.parametrize("case", data)


def pytest_runtest_setup(item):
    global modules
    global copy_modules

    env = item.config.getoption("--env")
    if len(modules) == 0:
        load_service_conf(env)

    envmarker = item.get_marker("module")

    if envmarker is not None:
        module = envmarker.args[0]
        if module in modules:
            try:
                copy_modules.remove(module)
            except ValueError:  # already removed or not in list
                pass
            pass
        else:
            pytest.skip("module is not in the env %r" % env)

    if item.function.__name__ == 'test_bft':
        exclude_list = exclude_dict.get(env, list())
        for e in exclude_list:
            generated_arg = 'case%d' % e
            if generated_arg == item._genid:
                pytest.skip('%s is in exclude list: %s' % (item._genid, exclude_list))


@pytest.fixture(scope='session', autouse=True)
def host(request):
    global _host
    _host = u'http://%s' % request.config.getoption('--host')
    return u'http://%s' % request.config.getoption('--host')


@pytest.fixture(scope='session', autouse=True)
def solr_host(request):
    host = request.config.getoption('--solr_host')
    if not host:
        host = request.config.getoption('--host')
    return u'http://%s' % host


def load_service_conf(env):
    global modules
    global copy_modules
    modules = []
    modules.append('solr')
    for i in service_conf.conf["services"]:
        for j in i["envs"]:
            if j["name"] == env:
                service_name = i["service_name"]
                modules.append(service_name.replace("-", ""))
    copy_modules = copy.deepcopy(modules)


# @pytest.fixture(scope='session')
# def chatlog_db(request):
#     user = request.config.getoption('--db_user')
#     password = request.config.getoption('--db_pass')
#     url = 'mysql+pymysql://%s:%s@%s:3306/backend_log?charset=utf8' % (
#         user, password, request.config.getoption('--db_url'))
#     e = create_engine(url, pool_recycle=2, pool_timeout=5, echo=True)
#     assert e.dialect.has_table(e, 'chat_record')
#     return e


@pytest.fixture(scope='session')
def ssh_conn(request):
    url = request.config.getoption('--host')
    username = request.config.getoption('--ssh_user')
    password = request.config.getoption('--ssh_pass')
    conn = SSHClient()
    conn.set_missing_host_key_policy(AutoAddPolicy)
    conn.connect(url, username=username,
                 password=password, look_for_keys=False)
    return conn


@pytest.fixture(scope='session')
def env(request):
    return request.config.getoption('--env')
