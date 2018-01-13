#!/bin/bash
project=suning
channel=mychannel
docker_ca_images="hyperledger/fabric-ca:x86_64-1.0.5"
docker_tool_imagers="hyperledger/fabric-tools:x86_64-1.0.5"
docker_peer_images="hyperledger/fabric-peer:x86_64-1.0.5"
docker_kafka_images="hyperledger/fabric-kafka:x86_64-1.0.5"
docker_zookpeer_images="hyperledger/fabric-zookeeper:x86_64-1.0.5"
docker_couchdb_images="hyperledger/fabric-couchdb:x86_64-1.0.5"
docker_orderer_images="docker.io/hyperledger/fabric-orderer:x86_64-1.0.5"
docker_ccenv_images="docker.io/hyperledger/fabric-ccenv:x86_64-1.0.5"
docker_baseos_images="docker.io/hyperledger/fabric-baseos:x86_64-0.4.2"

rm -rf "$project"_*

function create_ca() {
    dir="$project"_ca
    rm -rf $dir
    mkdir -p $dir

    echo "docker pull $docker_ca_images" >> $dir/download-dockerimages.sh
    echo "docker tag $docker_ca_images hyperledger/fabric-ca" >> $dir/download-dockerimages.sh
    echo "CHANNEL_NAME=$channel ./docker-compose -f docker-compose-ca.yaml up -d " >> $dir/start.sh
    echo "sleep 2" >> $dir/start.sh
#    echo "curl -X PUT http://127.0.0.1:5984/_users" >> $dir/start.sh
#    echo "curl -X PUT http://127.0.0.1:5984/_replicator" >> $dir/start.sh
#    echo "curl -X PUT http://127.0.0.1:5984/_global_changes" >> $dir/start.sh
    echo "docker logs -f ca_peerOrg1" >> $dir/start.sh
    echo "docker logs -f ca_peerOrg2" >> $dir/start.sh
	echo "docker logs -f ca_peerOrg3" >> $dir/start.sh
    echo "docker logs -f ca_peerOrg4" >> $dir/start.sh
    chmod u+x $dir/download-dockerimages.sh
    chmod u+x $dir/start.sh

    mkdir -p $dir/scripts
    cp -rf solo $dir
    cp -rf kafka $dir
    cp -rf scripts $dir
    cp -rf docker-compose $dir
    cp -rf docker-compose-ca.yaml $dir

    tar -cvf "$dir".tar $dir
    gzip "$dir".tar
    rm -rf $dir
}

function create_orderer() {

    dir="$project"_orderer
    rm -rf $dir
    mkdir -p $dir
    echo "docker pull $docker_orderer_images" >> $dir/download-dockerimages.sh
    echo "docker pull $docker_kafka_images" >> $dir/download-dockerimages.sh
    echo "docker pull $docker_zookpeer_images" >> $dir/download-dockerimages.sh
    echo "docker tag $docker_orderer_images hyperledger/fabric-orderer" >> $dir/download-dockerimages.sh
    echo "docker tag $docker_kafka_images hyperledger/fabric-kafka" >> $dir/download-dockerimages.sh
    echo "docker tag $docker_zookpeer_images hyperledger/fabric-zookeeper" >> $dir/download-dockerimages.sh
   #	   echo "CHANNEL_NAME=$channel ./docker-compose -f docker-compose-orderer-solo.yaml up -d " >> $dir/startSolo.sh
    echo "CHANNEL_NAME=$sunningchannel ./docker-compose -f docker-compose-orderer-kafka.yaml up -d " >> $dir/startKafka.sh
   #   echo "docker logs -f orderer.example.com" >> $dir/startSolo.sh
    echo "docker logs -f orderer.example.com" >> $dir/startKafka.sh
    chmod u+x $dir/download-dockerimages.sh
#chmod u+x $dir/startSolo.sh
    chmod u+x $dir/startKafka.sh
	
#    cp -rf solo $dir
    cp -rf kafka $dir
    cp -rf docker-compose $dir
#    cp -rf docker-compose-orderer-solo.yaml $dir
    cp -rf docker-compose-orderer-kafka.yaml $dir
    cp -rf peer-base $dir
    tar -cvf "$dir"0.tar $dir
    gzip "$dir"0.tar
    rm -rf $dir
}

function create_peer() {
    rm -rf $project
    mkdir -p $project

    for N in 0 1 2 3 ; do
        rm -rf $project
        mkdir -p $project
        echo "CHANNEL_NAME=$sunningchannel ./docker-compose -f docker-compose-peer"$N".yaml up -d " >> $project/start.sh
        echo "sleep 12" >> $project/start.sh
        echo "curl -X PUT http://127.0.0.1:5984/_users" >> $project/start.sh
        echo "curl -X PUT http://127.0.0.1:5984/_replicator" >> $project/start.sh
        echo "curl -X PUT http://127.0.0.1:5984/_global_changes" >> $project/start.sh
        echo "docker pull $docker_peer_images" >> $project/download-dockerimages.sh
        echo "docker tag $docker_peer_images hyperledger/fabric-peer" >> $project/download-dockerimages.sh
        echo "docker pull $docker_couchdb_images" >> $project/download-dockerimages.sh
        echo "docker tag $docker_couchdb_images hyperledger/fabric-couchdb" >> $project/download-dockerimages.sh
        echo "docker pull $docker_ccenv_images" >> $project/download-dockerimages.sh
        echo "docker tag $docker_ccenv_images hyperledger/fabric-ccenv" >> $project/download-dockerimages.sh
        echo "docker pull $docker_baseos_images" >> $project/download-dockerimages.sh
        chmod u+x $project/start.sh
        chmod u+x $project/download-dockerimages.sh

        mkdir -p ./$project/peer-base
        cp -rf docker-compose $project
        cp -rf docker-compose-peer"$N".yaml $project
        cp -rf peer-base/peer-base.yaml $project/peer-base/peer-base.yaml
        #cp -rf solo $project/
        cp -rf kafka $project/
		cp -rf ../src/ $project/
        tar -cvf "$project"_peer$N.tar $project
        gzip "$project"_peer$N.tar
        rm -rf $project
    done

    rm -rf $project
}


create_orderer

create_ca

create_peer

