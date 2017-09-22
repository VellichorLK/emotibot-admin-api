#!/usr/bin/python

import errno
import shutil
import sys
import os
import subprocess
import argparse


def _create_folder(folder):
    # real_path = os.path.realpath(output_folder)
    if os.path.exists(folder):
        if not os.path.isdir(folder):
            print '%s is not a folder!' % (folder)
            sys.exit(1)
    else:
        try:
            os.makedirs(folder)
        except Exception as exp:
            print 'makedirs %s failed: %s' % (folder, exp)
            sys.exit(1)
    return os.path.realpath(folder)


def do_save(output_folder, compose_file):
    real_path = _create_folder(output_folder)
    # docker-compose -f docker-compose.yml config | grep image
    ps = subprocess.Popen(['docker-compose', '-f', compose_file, 'config'],
                          stdout=subprocess.PIPE)
    raw_ret = subprocess.check_output(['grep', 'image'], stdin=ps.stdout)
    ret_lines = raw_ret.split('\n')
    for line in ret_lines:
        if not line:
            continue
        line = line.split("image:")[1]
        line = line.strip()
        line = line.lstrip()
        # print line
        name = line.rsplit(":", 1)[0]
        dst = '%s.tar' % (os.path.join(real_path, os.path.basename(name)))
        cmd = "docker save %s -o %s" % (line, dst)
        print cmd
        subprocess.call(cmd.split())


def do_load(folder):
    if not os.path.isdir(folder):
        print '%s is not dir' % folder
        sys.exit(1)

    for f in os.listdir(folder):
        cmd = "docker load -i %s" % os.path.join(folder, f)
        print cmd
        subprocess.call(cmd.split(" "))


def do_destroy(compose_file):
    # Delete all containers
    # docker rm $(docker ps -a -q)
    # Delete all images
    # docker rmi $(docker images -q)
    cmd = 'docker-compose -f %s rm -sf' % compose_file
    print cmd
    subprocess.call(cmd.split())
    ret = subprocess.check_output(['docker', 'images', '-q'])
    rets = ret.split('\n')
    for r in rets:
        if not r:
            continue
        cmd = 'docker rmi %s' % r
        subprocess.call(cmd.split())


def do_run(compose_file, env_file, services, depends, number):
    '''
    1) copy env_file to .env
    2) compose comand:
        docker-compose -f ./docker-compose.yml rm -s ${service}
        docker-compose -f ./docker-compose.yml up --force-recreate --remove-orphans ${depends} -d ${scale} ${service}
    '''
    # mkdir for test env
    if env_file.endswith('test.env'):
        try:
            os.makedirs('/tmp/persistant_storage')
        except Exception as exp:
            if exp.errno != errno.EEXIST:
                print 'makedirs %s failed: %s' % ('/tmp/persistant_storage',
                                                  exp)
                sys.exit(1)

    # copy env_file to .env
    dst_file = os.path.join(os.path.dirname(env_file), '.env')
    try:
        shutil.copyfile(env_file, dst_file)
    except Exception as exp:
        print 'copy %s to %s failed due to %s' % (env_file, dst_file, exp)
        sys.exit(1)

    # compose command: remove previous service
    cmd = 'docker-compose -f %s rm -sf %s' % (
        compose_file, ' '.join(n for n in services) if services else '')
    print '### exec cmd: [%s]' % cmd.strip()
    subprocess.call(cmd.strip().split(" "))

    # TODO: deal with depends and scale
    no_deps = ''
    scale = ''
    if services:
        if depends is False:
            no_deps = '--no-deps '
        for s in services:
            if s == 'worker-voice-emotion-analysis':
                scale = '--scale %s=%s ' % (s, number)
    cmd = 'docker-compose -f %s up --force-recreate --remove-orphans %s-d %s%s' % (
        compose_file, no_deps, scale, ' '.join(n for n in services) if services else '')
    print '### exec cmd: [%s]' % cmd.strip()
    subprocess.call(cmd.strip().split(" "))


def do_stop(compose_file, services):
    # docker-compose -f docker-compose.yml stop ${service}
    cmd = 'docker-compose -f %s stop %s' % (compose_file, ' '.join(n for n in services))
    print '### exec cmd: [%s]' % cmd.strip()
    subprocess.call(cmd.strip().split(" "))


def main():
    # parse args
    parser = argparse.ArgumentParser()
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument('--save', action='store_true', help='Save images. E.g. docker-compose --save -o ${/path/to/output_folder} -f ${compose_yaml}')
    group.add_argument('--load', action='store_true', help='Load images. E.g. docker-compose --load -o ${/path/to/output_folder}')
    group.add_argument('--destroy', action='store_true', help='Destrop ALL images. E.g. docker-compose --destroy -f ${compose_yaml}')
    group.add_argument('--run', action='store_true', help='Run service. E.g. docker-compose --run -f ${compose_yaml} -e ${env_file} -s ${service1} -s ${service2}')
    group.add_argument('--stop', action='store_true')
    parser.add_argument('-o', '--image_folder', default='/tmp/api_srv_images')
    parser.add_argument('-f', '--compose_file', default='./docker-compose.yml')
    parser.add_argument('-e', '--env', default='./test.env')
    parser.add_argument('-s', '--service', action='append', default=[])
    parser.add_argument('-d', '--depends', action='store_true', default=False,
                        help='if service is empty, depends always be true')
    parser.add_argument('-n', '--number', type=int, default=1,
                        help='only affect on voice analysis service')
    args = parser.parse_args()
    print args

    # do action
    if args.save:
        if not os.path.exists(args.compose_file):
            parser.print_help()
        do_save(args.image_folder, args.compose_file)
    elif args.load:
        do_load(args.image_folder)
    elif args.destroy:
        do_destroy(args.compose_file)
    elif args.run:
        if not os.path.exists(args.compose_file):
            parser.print_help()
        do_run(args.compose_file, args.env, args.service,
               args.depends, args.number)
    elif args.stop:
        if not os.path.exists(args.compose_file):
            parser.print_help()
        do_stop(args.compose_file, args.service)
    else:
        parser.print_help()


if __name__ == '__main__':
    main()
