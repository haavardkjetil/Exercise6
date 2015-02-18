package main

import (

"time"
"log"
"bytes"
"encoding/gob"
"sort"
"net"

func main() {  

	timer := time.NewTimer(time.Second)


	sendChan := make(chan int,10)
	receiveChan := make(chan int,10)
	quit := make(chan int, 2)

	go receive_message(receiveChan, quit)
	go send_message(sendChan, quit)

	teller := 0
	iAmPrimary := false
	for{
		select{
			case newpacket := <-receiveChan:
				if !iAmPrimay {
					teller = newpacket
					timer.Reset(time.Second)
				}

			case <- timer.C:
				if iAmPrimary {
					teller++
					println(teller)
						timer.Reset(time.Second)
				}else {
					iAmPrimary = true
				}
			default:
				if iAmPrimary {
					sendChan <- teller
				}
		}
		
	}
			
		
	}
	quit <- 1
	quit <- 1 
}


func receive_message(transmitChannel chan int, quit chan int) {
	//Initializing
	localAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort( "",udpRcvPort))
	if err != nil {
		log.Fatal( "Failed to resolve addr for :" + udpRcvPort, err );
	}

	recieveConnection, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatal("UDP recv connection error on " + localAddr.String(), err)
	}
	
	defer recieveConnection.Close()
	//Initialization done


		for {
			select{
				case <-quit:
					return
				default:
					// Decoderen må opprettes ny for hver pakke, fordi UDP benyttes og man
					// kan ikke være sikker på at den første pakken ble sendt (som inneholder type)
					receiveBufferRaw := make( []byte, 1600 ) // standard MTU size -- no packet should be bigger
					var receiveBuffer bytes.Buffer
					UDPpacketDecoder := gob.NewDecoder(&receiveBuffer)

					_, from, err := recieveConnection.ReadFromUDP( receiveBufferRaw )
					if from.String() == getMyIP() + ":" + udpSendPort {
						continue
					}
					if err != nil {
						log.Fatal("Error receiving UDP packet: " + err.Error(),err )
					}
					
					receiveBuffer.Write(receiveBufferRaw)
					mssg := Packet_t{} // Mulig feil her og neste linje
					err = UDPpacketDecoder.Decode(&mssg) 
					if err != nil {
						log.Print("Could not decode message: ", err)
						continue
					}
					transmitChannel <- mssg 
			}
		}
}

func send_message(transmitChannel chan int, quit chan int){
	remoteAddr, err := net.ResolveUDPAddr( "udp", net.JoinHostPort( bcast, udpRcvPort ) )
	if err != nil {
		log.Fatal("Failed to resolve UDP remote address:", err)
	}
	localAddr, err := net.ResolveUDPAddr( "udp", net.JoinHostPort("", udpSendPort ) )
	if err != nil {
		log.Fatal("Failed to resolve UDP local address:", err)
	}
	sendConnection, err := net.ListenUDP( "udp", localAddr)
	if err != nil {
		log.Fatal("UDP send connection error on " + localAddr.String() + ": ", err)
	}
	defer sendConnection.Close()

	for {
		select{
			case <-quit:
				return
			case newPacket := <- transmitChannel:
				
				var sendBuffer bytes.Buffer
				UDPpacketEncoder := gob.NewEncoder(&sendBuffer)
				sendBufferRaw := make( []byte, 1600)

				err = UDPpacketEncoder.Encode(newPacket)
				if err != nil {
					log.Fatal("Unable to encode new packet. ", err)
				}

				sendBuffer.Read(sendBufferRaw)
				_, err = sendConnection.WriteToUDP(sendBufferRaw, remoteAddr)
				if err != nil {
					log.Fatal("Failed to send packet from buffer: ", err)
				}
		}
	}
}
