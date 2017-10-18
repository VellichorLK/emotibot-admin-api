# -*- coding: utf-8 -*-
import os
import json
from datetime import datetime

import pytest
import requests
import healthcheck

from utils import backoff_request

_host = None


@pytest.fixture(scope='session', autouse=True)
def init_host(host):
    global _host
    if _host is None:
        _host = host


class TestVoiceEmotionAPI(object):
    def test_query_timerange_future():
        pass


class TestVoiceEmotionWorker(object):
    def test_non_ascii_filename():
        pass



def _compose_url(port, api, host=None):
    h = _host if host is None else host
    url = u'%s:%s' % (h, port)
    return os.path.join(url, api) if api != u'/' else url


def _check_val(target,  val):
    assert target == val, '%s != %s' % (target, val)


def _check_type(obj, type_):
    assert isinstance(obj, type_), '%s, %s, %s' % (obj, type(obj), type_)


def _check_keys(dict_obj, keys):
    _check_type(dict_obj, dict)
    if isinstance(keys, basestring):
        assert keys in dict_obj
    elif isinstance(keys, list):
        for k in keys:
            assert k in dict_obj, '%s not in %s' % (k, dict_obj.keys())


@pytest.fixture(scope='session')
def voice_test_data():
    path = os.path.join(os.path.dirname(__file__), 'data/6446.amr')
    return {
        'file': open(path, 'rb')
    }


@pytest.fixture(scope='session')
def intent_engine_data():
    path = os.path.join(os.path.dirname(__file__), 'data/intent_engine.txt')
    with open(path, 'rb') as f:
        return json.load(f)


@pytest.fixture(scope='session')
def rewrite_data():
    path = os.path.join(os.path.dirname(__file__), 'data/rewrite.txt')
    with open(path, 'rb') as f:
        return json.load(f)


@pytest.mark.module("emotibotwebcontroller")
class TestEmotibotWebController(object):
    @property
    def port(self):
        return 10901

    @property
    def api(self):
        return 'robot'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'UniqueID': 'test',
            'Text1': u'你好',
            'UserID': 'test'
        }
        resp = backoff_request(url=self.url, params=params, jsonize=True)
        _check_type(resp, dict)
        assert resp

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("wordtovecservice")
class TestWord2VecService(object):
    @property
    def port(self):
        return 11501

    @property
    def api(self):
        return 'qq_similarity'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'src': u'今天天气不错',
            'tar': u'你觉得今天天气怎么样'
        }
        resp = backoff_request(url=self.url, params=params)
        assert resp.status_code == requests.codes.ok

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("knowledgegraph")
class TestKnowledgeGraph(object):
    @property
    def port(self):
        return 11000

    @property
    def api(self):
        return 'json'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            't1': u'姚明的身高是多少',
            'robot': 'shadow',
            'knowledge_db': 'general'
        }
        resp = backoff_request(url=self.url, params=params, jsonize=True)
        _check_keys(resp, ['ver', 'answer'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("robotwriter")
class TestRobotWriter(object):
    '''Robotwriter module, under function module, offer weather/soccer/timeinfo
    '''
    @property
    def city_not_found(self):
        return u'海南'

    @property
    def port(self):
        return 10101

    def test_weather_normal(self, host):
        '''
        answer in format:
            {
                "answer": "[台北气温25℃~32℃，此刻26℃。中级风，注意尘土飞扬。哎哟，空气不错哦，木有一丝雾霾。],[明日天气，25~32度，中级风，风力2级。]",
                "statusCode": 200
            }
        '''
        params = {
            'city_name': '北京'
        }
        url = '%s:%s/V2/weather' % (host, self.port)
        ret = backoff_request(url=url, params=params, jsonize=True)

        assert isinstance(ret, dict)
        assert ret['statusCode'] == 200
        assert len(ret['answer'].split(',')) == 2

    def test_weather_not_support(self, host):
        '''Bugzilla XXXXX
        '''
        params = {
            'city_name': self.city_not_found
        }
        url = '%s:%s/V2/weather' % (host, self.port)
        ret = backoff_request(url=url, params=params, jsonize=True)
        assert ret['statusCode'] == 200
        assert u'暂未找到%s天气' % params['city_name'] in ret['answer']
        # assert False, '%s(%s)' % (ret['answer'], type(ret))

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("answer_classifier")
class TestAnswerClassifier(object):
    @property
    def port(self):
        return 10601

    @property
    def api(self):
        return 'qaScore'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'q': '1',
            'a': 'b',
            's': 'TIANYA'
        }
        resp = backoff_request(url=self.url, params=params)
        float(resp.text)
        # _check_type(resp.content, float)

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("topicbundle")
class TestTopicBudle(object):
    @property
    def port(self):
        return 10301

    @property
    def api(self):
        return ''

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        data = {
            "sentence": u"教練我想打球",
            "uniqueID": "REGRESSION_TEST"
        }
        resp = backoff_request(method='POST', url=self.url, json=data, jsonize=True)
        _check_keys(resp, 'results')
        _check_keys(resp['results'], ['version', 'candidated_ans', 'predictions', 'type'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("emotion")
class TestEmotion(object):
    @property
    def port(self):
        return 10401

    @property
    def api(self):
        return ''

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        data = {"UserId": "W165", "sentence": "你", "pos": [{"pos": "rr", "word": "你"}], "word_vector": [{"vector": [-4.908888, 6.1197596, 2.2497241, -2.1324494, 8.248826, 3.3449945, -7.1461387, 2.3615615, -5.998714, -6.1172037, 2.4235837, 1.4699548, -3.6426175, 0.26095083, 4.7892585, -3.6484928, 0.61304575, -2.6273496, 2.667583, -3.2750847, 3.2041593, 1.3548359, 2.2422945, -3.6373982, -2.5801735, -1.4217955, -2.8800511, -0.11498805, -1.34889, 3.5115023, 1.6250694, -2.8221478, -0.7345618, -1.9004483, 4.119014, 3.3368444, -5.1420374, -7.275899, 0.017379224, -3.3169518, -3.3542817, -2.7605846, 1.7251227, -5.3042946, -5.1701126, -0.9821165, 0.676945, 3.257999, -0.177063, -3.9779904, 2.1241758, -7.9531317, -1.0181433, -2.662671, -0.6606771, -0.11007068, 3.109022, -0.6054109, 4.2879963, -0.01754186, -0.23214437, -0.36398026, -1.7960869, -3.9192302, -3.3530242, 4.0164695, 3.3920896, 0.80102617, -4.7175183, -1.2997639, 2.986088, -7.564116, -1.641645, 3.3617718, 6.9458714, 3.9010904, 5.545931, 0.13412923, -0.27236575, 1.012459, 1.5605038, -1.415, 1.607941, -5.1526957, 2.881821, -2.723667, -1.5660717, -1.1938622, 3.564615, 3.8917055, -5.526187, 6.065005, -0.80775166, 1.1267289, -1.499784, 3.9274082, 1.5506396, 3.9550421, 0.3384558, 2.043788, -0.03228, 1.3362803, 2.4774718, 1.6162785, 4.193634, -2.2420366, 1.8727093, -4.272956, 0.6159371, -2.779468, -1.416059, -4.146396, 1.9420698, -1.1652635, 1.3417494, -1.781446, -2.390758, 2.2075663, -1.702454, 3.7517433, 0.7173769, 6.108675, 1.7748435, 4.615007, 3.9323308, 0.96440357, 1.9097383, 2.1827614, 3.3883302, 0.7073936, 5.207193, -2.4106903, 0.09921262, -9.375905, 7.3466835, -3.1209297, 2.4903026, 1.4238973, -3.0878825, -1.4413017, 3.5992076, 0.7948663, 4.2397842, -5.6235085, -0.66321856, 4.494518, -6.9030933, 3.5534153, -5.298325, 1.392322, -5.3320084, 3.9758105, -0.17044143, -3.092782, 1.7718723, -5.603978, 0.5350171, 3.2311819, -2.5765553, -1.0924823, -5.8809485, 5.1544642, -3.830129, 4.0784397, -1.068404, 1.7956712, -5.7058296, -3.0602062, 0.26532507, 0.6802529, 1.1895059, 1.3345408, -1.4302663, 1.845946, 1.1766301, 2.0739727, -2.0437717, -0.030036667, -2.7338333, -1.480569, 0.083817795, 0.051880147, -1.959884, 0.7635989, 1.4839134, -0.22905125, 3.3818154, 2.7380157, 4.5273585, -1.4694288, -0.9760939, 0.32054904, -2.3257017, 0.74233353, -4.41589, 1.013459, 0.99321514, 1.4849778, 1.6296649, 2.995613, 1.6818726, -7.550703, -3.5704606, 1.9867922, 3.8698258, -2.5225918, -2.4034524, -3.900074, -3.054884, -1.9020046, 0.12046142, 2.310737, 2.849426, 2.006298, -2.742948, -0.9596253, 4.646928, 3.2649958, -0.21117812, -1.7816002, -5.916985, -1.4087642, -4.018683, 3.3357916, 5.244464, -3.0034637, -0.6187805, 2.8751652, -1.1838403, -0.7559696, -0.9108595, -3.1147444, 3.471935, -2.1460505, -3.9395154, 2.6109734, 7.3127823, -1.9119477, 6.1358805, 1.1545424, -1.0435618, -1.8801502, 3.1248853, -3.6746597, 5.390668, 1.6625371, -4.588247, 3.5311553, -3.2505064, 5.9466467, -5.0726624, -6.530842, -3.7885222, -6.6497955, -6.340876, 0.938282, 6.433157, 7.7299247, -2.8431604, -0.083531335, -4.471328, 2.6202877, -1.1530069, 3.8771129, -2.3534427, -5.2725863, 0.6893261, 3.258497, -0.16690548, 1.3420979, -2.72647, -1.6055558, 3.5693228, -3.882773, -3.3298116, 1.7276219, 2.1139667, 2.612563, 5.220772, -1.4184622, -1.9691133, 3.802979, 0.26146898, 1.3973805, 0.2686918, -3.3498003, -0.38358566, 1.6009772, -0.8391515, 0.37769443, -3.3521814, 2.7023442, 3.1143804, -5.979715, -2.737778, 4.4697175, -3.75829, 1.3170402, 1.0061024, -3.4075296], "word": "你"}], "tokenized_sentence": ["你"]}
        resp = backoff_request(method='POST', url=self.url, json=data, jsonize=True)
        _check_keys(resp, 'results')
        _check_keys(resp['results'], ['type', 'version', 'candidated_ans', 'predictions'])

@pytest.mark.module("speechact")
class TestSpeechAct(object):
    @property
    def port(self):
        return 10201

    @property
    def api(self):
        return ''

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        data = {"sentences": "信件已经寄出"}
        resp = backoff_request(method='POST', url=self.url, json=data, jsonize=True)
        _check_keys(resp, 'results')
        _check_keys(resp['results'], 'message')
        assert resp['results']['message'] == 'Good'

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("solitaire")
class TestSolitaire(object):
    @property
    def port(self):
        return 12201

    @property
    def api(self):
        return 'api/idiom-solitaire'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        ret = backoff_request(method='POST', url=self.url,
                              json={"old": [], "input": "八拜之交"})
        assert ret.status_code == requests.codes.ok

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("voiceemotion")
class TestVoiceEmotion(object):
    @property
    def port(self):
        return 11801

    @property
    def api(self):
        return 'voice-emotion'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self, voice_test_data):
        data = {
            'type': 'woman'
        }
        resp = backoff_request(method='POST', url=self.url, data=data,
                               files=voice_test_data, jsonize=True)
        _check_keys(resp, ['ret', 'emotion'])

@pytest.mark.module("automaticcomposition")
class TestAutomaticComposition(object):
    @property
    def port(self):
        return 11802

    @property
    def api(self):
        return 'automatic-composition'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        data = {'type': 'sing', 'emotion': 'angry', 'userid': 55688, 'speed': 1}
        resp = backoff_request(method='POST', url=self.url, json=data, jsonize=True)
        _check_keys(resp, ['ret', 'url'])

@pytest.mark.module("houta")
class TestHouta(object):
    @property
    def port(self):
        return 80

    @property
    def api(self):
        return 'api/APP/chat2.php'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'text': u'上海天气',
            'wechatid': 'Test',
            'type': 'text'
        }
        resp = backoff_request(url=self.url, params=params, jsonize=True)
        assert resp['return'] == 0

@pytest.mark.module("cu")
class TestCU(object):
    @property
    def port(self):
        return 10701

    @property
    def api(self):
        return u'cu'

    def test_normal(self, host, env):
        url = '%s:%s/%s' % (host, self.port, self.api)
        params = {
            'UniqueID': u'123',
            'UserID': '100',
            'Text1': u'我不喜歡香蕉'
        }
        ret = backoff_request(url=url, params=params, jsonize=True)
        assert isinstance(ret, dict)
        columns = [
            'topic',
            'topic_mood',
            'speech_act',
            'wordPos',
            'keyWords',
            'Text0',
            'Text1',
            'Text1_Old',
            'UniqueID',
            'UserID']
        if env == 'vip':
            columns.remove('topic')
            columns.remove('topic_mood')
        for column in columns:
            assert column in ret

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("cuservice")
class TestCUService(object):
    @property
    def port(self):
        return 10801

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("responsecontroller")
class TestResponseController(object):
    @property
    def port(self):
        return 11601

    @property
    def api(self):
        return '/'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    @property
    def base_data(self):
        test_template = {"emotion":{"res":[{"item":"Neutral","score":100}],"ver":"20160330_1"},"topic":{"res":[],"ver":"20160330_1"},"speech_act":{"res":[{"item":"answer","score":0.2956872}],"ver":"20160330_1"},"wordPos":[],"keyWords":[],"keyWordsRule":[],"memory":[{"DDDD":123,"subject":"ffs", "relation":"喜歡", "entity":"姚明", "isConflicted":False}],"Text1":"","UniqueID":"123", "UserID":"1"}
        test_template["createdtime"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        return test_template

    def test_weather(self):
        data = self.base_data
        data.update({
            'Text1': u'今天会下雨吗'
        })

        resp = backoff_request(method='POST', url=self.url,
                               json=data,
                               jsonize=True)
        _check_keys(resp, ['score', 'emotion', 'answer', 'module', 'source', 'Text1'])

    def test_inference(self):
        data = self.base_data
        data.update({
            'Text1': u'明天早上八点叫我起床哦'
        })
        resp = backoff_request(method='POST', url=self.url,
                               json=data,
                               jsonize=True)
        _check_keys(resp, ['answer', 'module', 'score', 'Text1'])

    def test_hugry(self):
        data = self.base_data
        data.update({
            'Text1': u'我好饿啊'
        })
        resp = backoff_request(method='POST', url=self.url,
                               json=data,
                               jsonize=True)
        _check_keys(resp, ['answer', 'module', 'score', 'Text1'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("searchsong")
class TestSearchSong(object):
    @property
    def port(self):
        return 13201

    @property
    def api(self):
        return 'search/music'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url,
                               params={'name': u'一千首伤心的理由'},
                               jsonize=True)
        _check_keys(resp, ['baidu_Song', 'full_Match', 'search_Song', 'singer'])

@pytest.mark.module("memory")
class TestMemory(object):
    @property
    def port(self):
        return 11201

    @property
    def process_api(self):
        return u'memory/rest/process/post'

    @property
    def query_api(self):
        return u'memory/rest/query/get'

    def test_normal_get(self, host):
        params = {
            'type': 'scenario',
            'operation': 'query',
            'userId': 'W45',
            'invoker': 'memory'
        }
        url = u'%s:%s/%s' % (host, self.port, self.query_api)
        ret = backoff_request(url=url, params=params)
        assert ret.status_code == 200

    @pytest.mark.skip(reason='health check example')
    def test_normal_post(self, host):

        data = {
            "sentence":"有美即的口红吗",
            "rewriteSentence":"有美即的口红吗",
            "participles":[
                "有",
                "美即",
                "的",
                "口红",
                "吗"
            ],
            "userId":"101",
            "uniqueId":"123",
            "wordPos":[
                {
                    "word":"有",
                    "pos":"vyou"
                },
                {
                    "word":"美即",
                    "pos":"nz"
                },
                {
                    "word":"的",
                    "pos":"ude1"
                },
                {
                    "word":"口红",
                    "pos":"n"
                },
                {
                    "word":"吗",
                    "pos":"y"
                }
            ],
            "wordPosAfterRewrite":[
                {
                    "word":"有",
                    "pos":"vyou"
                },
                {
                    "word":"美即",
                    "pos":"nz"
                },
                {
                    "word":"的",
                    "pos":"ude1"
                },
                {
                    "word":"口红",
                    "pos":"n"
                },
                {
                    "word":"吗",
                    "pos":"y"
                }
            ],
        "conversationFeatures":{
            "topic_mood":"中性",
            "speech_act":"question-opinion",
            "topic":"化妆",
            "speech_act_confidence":"85.0",
            "intent_zoo":"其它"
        },
        "nlpFeatures":{
            "nlpRewrite":"有美即的口红吗",
            "person":"(PersonSubject:Other,PersonObject:Other)",
            "personNE":"",
            "namedEntities":"",
            "personStandard":"(PersonSubject:Other,PersonObject:Other)",
            "sentenceType":"是非问句",
            "polarity":"Polarity:Positive"
        }
        }
        url = '%s:%s/%s' % (host, self.port, self.process_api)
        ret = backoff_request(method='POST', url=url, data=data)
        # assert ret.status_code == 200
        assert False, ret.content

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("memorycrf")
class TestMemoryCRF(object):
    @property
    def port(self):
        return 13601

    @property
    def api(self):
        return 'crf/nickname'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url, params={'q': u'我叫小竹子'})
        assert resp.status_code == 200

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)


class TestSolr(object):
    @property
    def port(self):
        return 8081

    @property
    def api(self):
        return 'solr/merge_6_25'

    def test_normal(self, solr_host):
        url = _compose_url(self.port, self.api, host=solr_host)
        resp = backoff_request(url=url)
        assert resp.status_code == requests.codes.not_found


@pytest.mark.module("semanticrolelabeler")
class TestSRL(object):
    @property
    def port(self):
        return 12401

    @property
    def api(self):
        return 'SRL'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'text': u'我想退货',
            'doTree': 'true'
        }
        resp = backoff_request(url=self.url, params=params, jsonize=True)
        _check_keys(resp, ['lang', 'sentence', 'tree', 'srl', 'tokens', 'root_verb', 'model', 'manual_case'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("topicmood")
class TestTopicMood(object):
    @property
    def port(self):
        return 12701

    @property
    def api(self):
        return 'mood'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        params = {
            'sent': u'你烦不烦？',
            'uniqueID': 'REGRESSIOM_TEST'
        }
        resp = backoff_request(url=self.url, params=params, jsonize=True)
        _check_keys(resp, ['predictions', 'version', 'type'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("fileservice")
class TestFileService(object):
    @property
    def port(self):
        return 13001

    @property
    def api(self):
        return 'resource'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url, params={'label': 'Common/Silly'}, jsonize=True)
        _check_keys(resp, ['tag', 'url', 'web_base_url'])

@pytest.mark.module("searchconcert")
class TestSearchConcert(object):
    @property
    def port(self):
        return 13701

    @property
    def api(self):
        return 'concertservice/search'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url, params={'performers': u'周杰伦'})
        assert resp.status_code == requests.codes.ok

@pytest.mark.module("statelessfunction")
class TestStatelessFunction(object):
    @property
    def port(self):
        return 12601

    @property
    def api(self):
        return 'horoscope/today'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url, params={'cons_name': u'白羊座'},
                               jsonize=True)
        _check_keys(resp, ['statusCode', 'answer'])

@pytest.mark.module("sentencetype")
class TestSentenceType(object):
    @property
    def port(self):
        return 13401

    @property
    def api(self):
        return 'getType'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url, params={'s': u'你是谁'})
        assert resp.status_code == requests.codes.ok

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("intent")
class TestIntent(object):
    @property
    def port(self):
        return 13401

    @property
    def api(self):
        return 'getType'

    @property
    def url(self):
        return _compose_url(self.port, self.api)


@pytest.mark.module("dialoguecontroller")
class TestDialogueController(object):
    @property
    def port(self):
        return 12801

    @property
    def api(self):
        return u'DC'

    def test_normal(self, host):
        url = '%s:%s/%s' % (host, self.port, self.api)
        data = {
            "emotion":
            {
                "res":[],
                "ver":"20160504_2"
            },
            "speech_act":
            {
                "res":
                [
                    {"item":"question-info","score":50.01407251},
                    {"item":"answer","score":0.3240819},
                    {"item":"performative","score":0.02},
                    {"item":"commit","score":0.3610494},
                    {"item":"statement","score":0.28079617}
                ],
                "ver":"20160504_2"
            },
            "topic":
            {
                "res":
                [
                    {"item":"boyfriend","score":80},
                    {"item":"hhhh","score":80}
                ],
                "ver":"20160504_2"
            },
            "wordPos":
            [
                {"word":"明天","pos":"t"},
                {"word":"早上","pos":"t"},
                {"word":"八","pos":"m"},
                {"word":"点","pos":"qt"},
                {"word":"叫","pos":"vi"},
                {"word":"我","pos":"rr"},
                {"word":"起床","pos":"vi"}
            ],
            "keyWords":
            [
                {"keyword":"明天","level":1},
                {"keyword":"早上","level":2},
                {"keyword":"叫","level":1},
                {"keyword":"起床","level":1}
            ],
            "keyWordsRule":
            [
                {"keywordRule":"明天","levelRule":1},
                {"keywordRule":"早上","levelRule":2},
                {"keywordRule":"叫","levelRule":1},
                {"keywordRule":"起床","levelRule":1}
            ],
        "Text0":"天气",
        "Text1":"天气",
        "Text1_Old":"天气",
        "UniqueID":"124027",
        "UserID":"W106"
        }
        ret = backoff_request(method='POST', url=url, data=data)
        assert ret.status_code == 200

@pytest.mark.module("latestrecommend")
class TestLatestRecommend(object):
    @property
    def port(self):
        return 12501

    @property
    def api(self):
        return 'hotnews'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url,
                               params={
                                   'date': '20160910',
                                   'field': u'电影'
                               },
                               jsonize=True)
        _check_keys(resp, ['data', 'date', 'field', 'subfield'])

@pytest.mark.module("cookbookcontent")
class TestCookbookContent(object):
    @property
    def port(self):
        return 14201

    @property
    def api(self):
        return 'cb/V1/find'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        resp = backoff_request(url=self.url,
                               params={
                                   'type': 'class',
                                   'words': u'鸡蛋,苦瓜'
                               },
                               jsonize=True)
        assert resp


@pytest.mark.module("intentzoo")
class TestIntentZoo(object):
    @property
    def port(self):
        return 14301

    @property
    def api(self):
        return u'intent'

    def test_normal(self, host):
        url = '%s:%s/%s' % (host, self.port, self.api)
        params = {
            'uniqueID': '-1',
            'sent': '今天的新闻'
        }
        ret = backoff_request(url=url, params=params, jsonize=True)
        for column in ['version', 'predictions', 'type']:
            assert column in ret.keys()

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("solretlagent")
class TestSolrEtlAgent(object):
    @property
    def port(self):
        return 14401

    @property
    def api(self):
        return u'/'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_normal(self):
        ret = backoff_request(url=self.url)
        _check_val(ret.status_code, requests.codes.ok)

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("multicustomer")
class TestMultiCustomer(object):
    @property
    def port(self):
        return 14501

    @property
    def api(self):
        return u'/'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def xtest_normal(self):
        ret = backoff_request(url=self.url)
        _check_val(ret.status_code, requests.codes.not_found)

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("snluweb")
class TestNLU(object):
    @property
    def port(self):
        return 13901

    @property
    def api(self):
        return u'/'

    @property
    def health_api(self):
        return '_health_check'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    @property
    def health_url(self):
        return _compose_url(self.port, self.health_api)

    def test_normal(self):
        resp = backoff_request(url=self.url,
                               params={
                                   'f': 'all',
                                   't': 'false',
                                   'appid': '1200',
                                   'q': u'我喜欢你'
                               },
                               jsonize=True)
        _check_type(resp, list)
        _check_type(resp[0], dict)
        _check_val(resp[0]['nlpState'], 0)

    def test_health_check(self):
        resp = backoff_request(url=self.health_url)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("intentengine")
class TestIntentEngine(object):
    @property
    def port(self):
        return 15001

    @property
    def api(self):
        return u'intent_engine/tagging'

    @property
    def health_api(self):
        return '_health_check'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    @property
    def health_url(self):
        return _compose_url(self.port, self.health_api)

    def test_normal(self, intent_engine_data):
        resp = backoff_request(method='POST', url=self.url, json=intent_engine_data)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("customcu")
class TestCustomCU(object):
    @property
    def port(self):
        return 15201

    @property
    def api(self):
        return u'custom_cu'

    def test_normal(self, host):
        url = '%s:%s/%s' % (host, self.port, self.api)
        params = {
            'userid': '1',
            'uuid': '2',
            'appid': '9517fd0bf8faa655990a4dffe358e13e',
            'sentence': u'吃牛肉面花了一百元'
        }
        ret = backoff_request(url=url, params=params, jsonize=True)
        assert isinstance(ret, list)
        for column in ["topic-mood", "type", "topic", "intent"]:
            assert column in ret[0]

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("rewrite")
class TestRewrite(object):
    @property
    def port(self):
        return 11301

    @property
    def api(self):
        return u'rewrite'

    @property
    def health_api(self):
        return '_health_check'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    @property
    def health_url(self):
        return _compose_url(self.port, self.health_api)

    def test_normal(self, rewrite_data):
        resp = backoff_request(method='POST', url=self.url, json=rewrite_data)
        _check_val(resp.status_code, requests.codes.ok)

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

@pytest.mark.module("botfactorybackendservice")
class TestBotFactoryBackendService(object):
    @property
    def port(self):
        return 15501

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)


@pytest.mark.module("custominferenceservice")
class TestCustomInferenceService(object):
    @property
    def port(self):
        return 15601

    @property
    def api(self):
        return 'classify'

    @property
    def url(self):
        return _compose_url(self.port, self.api)

    def test_similar(self):
        params = {
            'text': u'卖的好贵',
            'top': 2
        }

        resp = backoff_request(url=self.url, params=params, jsonize=True)
        _check_keys(resp, ['Status', 'Msg'])
        _check_val(resp['Status'], 'Success')
        assert len(resp['Msg']) == 2

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)


@pytest.mark.module('vipadapter')
class TestVipConvertor(object):
    @property
    def port(self):
        return 15801

    @property
    def q_api(self):
        return 'vip/irobot/get-questions.action'

    @property
    def ask_api(self):
        return 'vip/irobot/ask4Json'

    @property
    def q_url(self):
        return _compose_url(self.port, self.q_api)

    @property
    def ask_url(self):
        return _compose_url(self.port, self.ask_api)

    def test_hotQ(self):
        params = {
            'Mode': 1,
            'top': 6,
            'platform': 'web',
            'qtype': 'hotQuestions'
        }
        resp = backoff_request(url=self.q_url, params=params, jsonize=True)
        _check_type(resp, list)

    def xtest_relatedQ(self):
        params = {
            'question': u'送货时间',
            'top': 6,
            'platform': 'web',
            'qtype': 'relatedQuestions'
        }

        resp = backoff_request(url=self.q_url, params=params, jsonize=True)
        _check_type(resp, list)

    def test_suggestQ(self):
        params = {
            'input': u'图',
            'top': 6,
            'platform': 'web',
            'qtype': 'suggestedQuestions'
        }
        resp = backoff_request(url=self.q_url, params=params, jsonize=True)
        _check_type(resp, list)

    def test_qa(self):
        # userId=abc123456&platform=weixin&question=答案列表
        params = {
            'userId': 'abc12345',
            'platform': 'weixin',
            'question': u'答案列表'
        }

        resp = backoff_request(url=self.ask_url, params=params,
                               jsonize=True)
        _check_type(resp, dict)
        _check_keys(resp, ['content', 'nodeId', 'moduleId', 'similarity',
                           'type', 'relatedQuestions', 'standardQuestion'])

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)


class TestFunctionContent(object):
    pass


class TestCommonParserService(object):
    pass


class TestCtripParserService(object):
    pass


class TestSentenceEmbeddingSimilarity(object):
    pass


class TestContextGraph(object):
    pass

@pytest.mark.module("taskengine")
class TestTaskEngine(object):
    @property
    def port(self):
        return 14101

    def test_health_check(self, host):
        resp = healthcheck.check_health_link(host, self.port)
        _check_val(resp.status_code, requests.codes.ok)

class TestTaskParser(object):
    pass

